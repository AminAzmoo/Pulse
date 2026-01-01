package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
	"pulse/backend/pkg/external/logging"
)

const (
	defaultLightTimeoutSeconds = 30
	defaultHeavyTimeoutSeconds = 300
	UpdateKindStatus           = "status"
	UpdateKindCapacity         = "capacity"
	UpdateKindMetrics          = "metrics"
)

type TaskRequest struct {
	Identity string                 `json:"identity"`
	Type     string                 `json:"type"`
	Class    string                 `json:"class"` // "heavy" | "light"
	Params   map[string]interface{} `json:"params"`
}

type TaskEnvelope struct {
	TaskID         uint                   `json:"task_id"`
	Identity       string                 `json:"identity"`
	Type           string                 `json:"type"`
	Class          string                 `json:"class"`
	IdempotencyKey string                 `json:"idempotency_key"`
	TimeoutSeconds int                    `json:"timeout_seconds"`
	Params         map[string]interface{} `json:"params"`
}

type TaskStatusUpdate struct {
	Kind string `json:"kind"` // "status" | "capacity" | "metrics"
	// status
	TaskID uint   `json:"task_id,omitempty"`
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
	// capacity
	HeavyAvailable int `json:"heavy_available,omitempty"`
	LightAvailable int `json:"light_available,omitempty"`
}

type TaskService struct {
	db *gorm.DB
}

func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{db: db}
}

func (s *TaskService) computeIdempotencyKey(req TaskRequest) string {
	// scope: identity + type + params canonical JSON
	b, _ := json.Marshal(req.Params)
	h := sha256.Sum256([]byte(req.Identity + "|" + req.Type + "|" + string(b)))
	return hex.EncodeToString(h[:])
}

func (s *TaskService) ValidateTaskParams(taskType string, params map[string]interface{}) error {
	switch taskType {
	case "exec_command":
		if _, ok := params["command"].(string); !ok {
			return errors.New("missing or invalid 'command' parameter")
		}
	case "file_download":
		if _, ok := params["url"].(string); !ok {
			return errors.New("missing or invalid 'url' parameter")
		}
		if _, ok := params["dest"].(string); !ok {
			return errors.New("missing or invalid 'dest' parameter")
		}
	case "service_action":
		if _, ok := params["service"].(string); !ok {
			return errors.New("missing or invalid 'service' parameter")
		}
		action, ok := params["action"].(string)
		if !ok {
			return errors.New("missing or invalid 'action' parameter")
		}
		validActions := map[string]bool{"start": true, "stop": true, "restart": true, "status": true}
		if !validActions[action] {
			return errors.New("invalid action")
		}
	case "self_update":
		if _, ok := params["url"].(string); !ok {
			return errors.New("missing or invalid 'url' parameter")
		}
		if _, ok := params["sha256"].(string); !ok {
			return errors.New("missing or invalid 'sha256' parameter")
		}
	}
	return nil
}

func (s *TaskService) CreateTask(req TaskRequest) (models.Task, bool, error) {
	if req.Identity == "" || req.Type == "" || req.Class == "" {
		return models.Task{}, false, errors.New("missing required fields")
	}
	if req.Class != models.TaskClassHeavy && req.Class != models.TaskClassLight {
		return models.Task{}, false, errors.New("invalid task class")
	}

	if err := s.ValidateTaskParams(req.Type, req.Params); err != nil {
		return models.Task{}, false, err
	}

	idem := s.computeIdempotencyKey(req)

	// check existing idempotency record
	var existing models.Task
	if err := s.db.Where("idempotency_key = ?", idem).First(&existing).Error; err == nil {
		return existing, true, nil
	} else if err != gorm.ErrRecordNotFound {
		return models.Task{}, false, err
	}

	timeout := defaultLightTimeoutSeconds
	if req.Class == models.TaskClassHeavy {
		timeout = defaultHeavyTimeoutSeconds
	}

	payloadBytes, _ := json.Marshal(req.Params)
	task := models.Task{
		Identity:       req.Identity,
		Type:           req.Type,
		Class:          req.Class,
		IdempotencyKey: idem,
		Status:         models.TaskStatusPending,
		PayloadJSON:    string(payloadBytes),
		TimeoutSeconds: timeout,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := s.db.Create(&task).Error; err != nil {
		return models.Task{}, false, err
	}

	logging.Audit("Task created", map[string]interface{}{"task_id": task.ID, "identity": task.Identity, "type": task.Type, "class": task.Class})
	return task, false, nil
}

func (s *TaskService) Envelop(task models.Task) TaskEnvelope {
	var params map[string]interface{}
	_ = json.Unmarshal([]byte(task.PayloadJSON), &params)
	return TaskEnvelope{
		TaskID:         task.ID,
		Identity:       task.Identity,
		Type:           task.Type,
		Class:          task.Class,
		IdempotencyKey: task.IdempotencyKey,
		TimeoutSeconds: task.TimeoutSeconds,
		Params:         params,
	}
}

func (s *TaskService) NextTasksFor(identity string, heavySlots, lightSlots int) ([]TaskEnvelope, error) {
	var tasks []models.Task
	// heavy first, then light, respect available slots
	if heavySlots > 0 {
		var heavy []models.Task
		if err := s.db.Where("identity = ? AND class = ? AND status = ?", identity, models.TaskClassHeavy, models.TaskStatusPending).Order("created_at asc").Limit(heavySlots).Find(&heavy).Error; err != nil {
			return nil, err
		}
		tasks = append(tasks, heavy...)
	}
	if lightSlots > 0 {
		var light []models.Task
		if err := s.db.Where("identity = ? AND class = ? AND status = ?", identity, models.TaskClassLight, models.TaskStatusPending).Order("created_at asc").Limit(lightSlots).Find(&light).Error; err != nil {
			return nil, err
		}
		tasks = append(tasks, light...)
	}
	envelopes := make([]TaskEnvelope, 0, len(tasks))
	for _, t := range tasks {
		envelopes = append(envelopes, s.Envelop(t))
	}
	return envelopes, nil
}

func (s *TaskService) MarkInProgress(taskID uint) error {
	return s.db.Model(&models.Task{}).Where("id = ? AND status = ?", taskID, models.TaskStatusPending).Updates(map[string]interface{}{
		"status":     models.TaskStatusInProgress,
		"updated_at": time.Now().UTC(),
	}).Error
}

func (s *TaskService) HandleStatusUpdate(update TaskStatusUpdate) error {
	if update.Kind == UpdateKindCapacity {
		// capacity updates are not persisted as tasks; they are used by gRPC loop
		return nil
	}
	if update.Kind == UpdateKindMetrics {
		return nil
	}
	if update.Kind != UpdateKindStatus {
		return errors.New("unsupported update kind")
	}
	var task models.Task
	if err := s.db.First(&task, update.TaskID).Error; err != nil {
		return err
	}
	// Only advance to completed/failed based on agent events
	switch update.Status {
	case models.TaskStatusAwaitingConfirmation, models.TaskStatusInProgress, models.TaskStatusFailed, models.TaskStatusCompleted:
		// allowed transitions
	default:
		return errors.New("invalid status transition")
	}
	if err := s.db.Model(&models.Task{}).Where("id = ?", update.TaskID).Updates(map[string]interface{}{
		"status":     update.Status,
		"updated_at": time.Now().UTC(),
	}).Error; err != nil {
		return err
	}

	// Update idempotency record
	outcome := map[string]interface{}{"task_id": update.TaskID, "status": update.Status}
	outcomeJSON, _ := json.Marshal(outcome)
	rec := models.IdempotencyRecord{
		IdempotencyKey:  task.IdempotencyKey,
		LastStatus:      update.Status,
		LastOutcomeJSON: string(outcomeJSON),
		UpdatedAt:       time.Now().UTC(),
	}
	if err := s.db.Where(models.IdempotencyRecord{IdempotencyKey: task.IdempotencyKey}).Assign(rec).FirstOrCreate(&rec).Error; err != nil {
		// Fallback: upsert by manual
		_ = s.db.Where("idempotency_key = ?", task.IdempotencyKey).Delete(&models.IdempotencyRecord{}).Error
		_ = s.db.Create(&rec).Error
	}

	logging.Audit("Task status updated", map[string]interface{}{"task_id": update.TaskID, "status": update.Status})
	return nil
}

func (s *TaskService) ListTasks(identity string, limit int) ([]models.Task, error) {
	var tasks []models.Task
	q := s.db.Order("created_at desc")
	if identity != "" {
		q = q.Where("identity = ?", identity)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) GetTaskByID(id uint) (models.Task, error) {
	var task models.Task
	if err := s.db.First(&task, id).Error; err != nil {
		return models.Task{}, err
	}
	return task, nil
}
