package services

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
	"pulse/backend/pkg/external/logging"
)

type IdentityService struct {
	db       *gorm.DB
	notifier *RevocationNotifier
	alertSvc *AlertService
}

func NewIdentityService(db *gorm.DB) *IdentityService {
	return &IdentityService{
		db:       db,
		notifier: NewRevocationNotifier(),
		alertSvc: NewAlertService(db),
	}
}

func (s *IdentityService) IsAllowed(identity string) (bool, error) {
	var node models.Node
	if err := s.db.Where("identity = ? AND status = ?", identity, models.NodeStatusActive).First(&node).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		logging.Logger.Error("Failed to check identity allowance", zap.Error(err), zap.String("identity", identity))
		return false, err
	}
	return true, nil
}

func (s *IdentityService) Revoke(identity string) error {
	var node models.Node
	if err := s.db.Where("identity = ?", identity).First(&node).Error; err != nil {
		return err
	}

	node.Status = models.NodeStatusRevoked
	if err := s.db.Save(&node).Error; err != nil {
		return err
	}

	s.notifier.Notify(identity)

	metadata := map[string]interface{}{"identity": identity}
	logging.Audit("Identity revoked", metadata)

	return nil
}

func (s *IdentityService) CreateNode(identity string) (models.Node, error) {
	node := models.Node{Identity: identity, Status: models.NodeStatusActive}
	if err := s.db.Create(&node).Error; err != nil {
		return models.Node{}, err
	}
	logging.Audit("Node created", map[string]interface{}{"node_id": node.ID, "identity": identity})
	return node, nil
}

func (s *IdentityService) RevokeNode(id uint) error {
	var node models.Node
	if err := s.db.First(&node, id).Error; err != nil {
		return err
	}
	node.Status = models.NodeStatusRevoked
	if err := s.db.Save(&node).Error; err != nil {
		return err
	}
	s.notifier.Notify(node.Identity)
	logging.Audit("Node revoked", map[string]interface{}{"node_id": node.ID})
	return nil
}

func (s *IdentityService) GetActiveNodeCount() (int, error) {
	var count int64
	if err := s.db.Model(&models.Node{}).Where("status = ?", models.NodeStatusActive).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (s *IdentityService) GetRevokedNodeCount() (int, error) {
	var count int64
	if err := s.db.Model(&models.Node{}).Where("status = ?", models.NodeStatusRevoked).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (s *IdentityService) GetTotalActiveAlerts() (int, error) {
	count, err := s.alertSvc.GetTotalActiveAlerts()
	return int(count), err
}

func (s *IdentityService) Subscribe(identity string) chan struct{} {
	return s.notifier.Subscribe(identity)
}

func (s *IdentityService) Unsubscribe(identity string, ch chan struct{}) {
	s.notifier.Unsubscribe(identity, ch)
}

func (s *IdentityService) UpdateHostInfo(identity string, info *models.Node) error {
	var node models.Node
	if err := s.db.Where("identity = ?", identity).First(&node).Error; err != nil {
		return err
	}

	node.Hostname = info.Hostname
	node.OS = info.OS
	node.Platform = info.Platform
	node.PlatformFamily = info.PlatformFamily
	node.PlatformVersion = info.PlatformVersion
	node.KernelVersion = info.KernelVersion
	node.Arch = info.Arch
	node.AgentVersion = info.AgentVersion
	node.UptimeSeconds = info.UptimeSeconds
	node.IPAddresses = info.IPAddresses
	node.LastSeenAt = time.Now()

	return s.db.Save(&node).Error
}

func (s *IdentityService) RecordMetrics(identity string, metrics *models.Metric) error {
	var node models.Node
	if err := s.db.Where("identity = ?", identity).Select("id").First(&node).Error; err != nil {
		return err
	}
	metrics.NodeID = node.ID
	metrics.Timestamp = time.Now()

	if err := s.db.Create(metrics).Error; err != nil {
		return err
	}

	// Trigger alert evaluation
	go s.alertSvc.Evaluate(node.ID, metrics)

	return nil
}

func (s *IdentityService) GetNode(id uint) (models.Node, error) {
	var node models.Node
	if err := s.db.First(&node, id).Error; err != nil {
		return models.Node{}, err
	}
	return node, nil
}

func (s *IdentityService) GetNodeMetrics(nodeID uint, limit int) ([]models.Metric, error) {
	var metrics []models.Metric
	if err := s.db.Where("node_id = ?", nodeID).Order("timestamp desc").Limit(limit).Find(&metrics).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}

func (s *IdentityService) ListNodes() ([]models.Node, error) {
	var nodes []models.Node
	if err := s.db.Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func (s *IdentityService) ListEvents(limit int) ([]models.Event, error) {
	var events []models.Event
	if err := s.db.Order("created_at desc").Limit(limit).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

type RevocationNotifier struct {
	mu   sync.RWMutex
	subs map[string][]chan struct{}
}

func NewRevocationNotifier() *RevocationNotifier {
	return &RevocationNotifier{
		subs: make(map[string][]chan struct{}),
	}
}

func (n *RevocationNotifier) Subscribe(identity string) chan struct{} {
	ch := make(chan struct{})
	n.mu.Lock()
	n.subs[identity] = append(n.subs[identity], ch)
	n.mu.Unlock()
	return ch
}

func (n *RevocationNotifier) Unsubscribe(identity string, ch chan struct{}) {
	n.mu.Lock()
	defer n.mu.Unlock()

	chans, ok := n.subs[identity]
	if !ok {
		return
	}

	for i, c := range chans {
		if c == ch {
			chans = append(chans[:i], chans[i+1:]...)
			break
		}
	}

	if len(chans) == 0 {
		delete(n.subs, identity)
	} else {
		n.subs[identity] = chans
	}
}

func (n *RevocationNotifier) Notify(identity string) {
	n.mu.RLock()
	chans, ok := n.subs[identity]
	n.mu.RUnlock()
	if !ok {
		return
	}

	for _, ch := range chans {
		close(ch)
	}

	n.mu.Lock()
	delete(n.subs, identity)
	n.mu.Unlock()
}
