package services

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
	"pulse/backend/pkg/external/logging"
)

type AlertService struct {
	db *gorm.DB
}

func NewAlertService(db *gorm.DB) *AlertService {
	// Ensure default rules exist
	svc := &AlertService{db: db}
	svc.seedDefaultRules()
	return svc
}

func (s *AlertService) seedDefaultRules() {
	rules := []models.AlertRule{
		{Name: "High CPU", Metric: "cpu", Operator: ">", Threshold: 90.0, Severity: "critical"},
		{Name: "High Memory", Metric: "memory", Operator: ">", Threshold: 90.0, Severity: "warning"},
		{Name: "Disk Full", Metric: "disk", Operator: ">", Threshold: 95.0, Severity: "critical"},
	}

	for _, r := range rules {
		var count int64
		s.db.Model(&models.AlertRule{}).Where("name = ?", r.Name).Count(&count)
		if count == 0 {
			s.db.Create(&r)
		}
	}
}

func (s *AlertService) Evaluate(nodeID uint, metrics *models.Metric) {
	var rules []models.AlertRule
	if err := s.db.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		logging.Logger.Error("Failed to load alert rules", zap.Error(err))
		return
	}

	for _, rule := range rules {
		val := s.getValue(metrics, rule.Metric)
		firing := s.checkCondition(val, rule.Operator, rule.Threshold)

		if firing {
			s.fireAlert(nodeID, rule, val)
		} else {
			s.resolveAlert(nodeID, rule)
		}
	}
}

func (s *AlertService) getValue(m *models.Metric, metricName string) float64 {
	switch metricName {
	case "cpu":
		return m.CPU
	case "memory":
		return m.Memory
	case "disk":
		return m.Disk
	default:
		return 0
	}
}

func (s *AlertService) checkCondition(val float64, op string, threshold float64) bool {
	switch op {
	case ">":
		return val > threshold
	case ">=":
		return val >= threshold
	case "<":
		return val < threshold
	case "<=":
		return val <= threshold
	default:
		return false
	}
}

func (s *AlertService) fireAlert(nodeID uint, rule models.AlertRule, val float64) {
	// Check if already firing
	var existing models.Alert
	// Check for active alerts (firing)
	// NOTE: We preload Rule here to ensure we don't duplicate logic, but strictly we just need existence
	err := s.db.Where("node_id = ? AND rule_id = ? AND status = ?", nodeID, rule.ID, "firing").First(&existing).Error

	if err == nil {
		// Update value
		existing.Value = val
		existing.TriggeredAt = time.Now() // Bump timestamp to show it's still active
		s.db.Save(&existing)
		return
	}

	// Create new alert
	alert := models.Alert{
		NodeID:      nodeID,
		RuleID:      rule.ID,
		Value:       val,
		Status:      "firing",
		TriggeredAt: time.Now(),
	}
	if err := s.db.Create(&alert).Error; err == nil {
		logging.Logger.Warn("Alert Fired",
			zap.Uint("node_id", nodeID),
			zap.String("rule", rule.Name),
			zap.Float64("value", val),
		)
		// TODO: Send notification (email/slack)
	}
}

func (s *AlertService) resolveAlert(nodeID uint, rule models.AlertRule) {
	var existing models.Alert
	err := s.db.Where("node_id = ? AND rule_id = ? AND status = ?", nodeID, rule.ID, "firing").First(&existing).Error
	if err != nil {
		return // No active alert
	}

	now := time.Now()
	existing.Status = "resolved"
	existing.ResolvedAt = &now
	if err := s.db.Save(&existing).Error; err == nil {
		logging.Logger.Info("Alert Resolved",
			zap.Uint("node_id", nodeID),
			zap.String("rule", rule.Name),
		)
	}
}

func (s *AlertService) CreateRule(rule models.AlertRule) (models.AlertRule, error) {
	if err := s.db.Create(&rule).Error; err != nil {
		return models.AlertRule{}, err
	}
	return rule, nil
}

func (s *AlertService) GetActiveAlerts(nodeID uint) ([]models.Alert, error) {
	var alerts []models.Alert
	q := s.db.Preload("Rule").Where("status = ?", "firing")
	if nodeID != 0 {
		q = q.Where("node_id = ?", nodeID)
	}
	if err := q.Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

func (s *AlertService) GetTotalActiveAlerts() (int64, error) {
	var count int64
	if err := s.db.Model(&models.Alert{}).Where("status = ?", "firing").Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
