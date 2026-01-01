package api

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"pulse/backend/internal/services"
	"pulse/backend/pkg/external/logging"
)

const (
	MaxNodeNameLength = 255
	SettingsCacheTTL  = 300 // seconds
)

func SetupRoutes(app *fiber.App, authMiddleware fiber.Handler, identitySvc *services.IdentityService, taskSvc *services.TaskService) {
	v1 := app.Group("/v1", authMiddleware)

	v1.Post("/nodes", createNodeHandler(identitySvc))
	v1.Post("/nodes/:id/token", generateTokenHandler(identitySvc))
	v1.Post("/nodes/:id/revoke", revokeNodeHandler(identitySvc))
	v1.Post("/nodes/:id/tasks", createNodeTaskHandler(identitySvc, taskSvc))
	v1.Get("/nodes", listNodesHandler(identitySvc))
	v1.Get("/nodes/:id", getNodeHandler(identitySvc))
	v1.Get("/nodes/:id/metrics", getNodeMetricsHandler(identitySvc))
	v1.Get("/settings", settingsHandler(identitySvc))
	v1.Get("/events", listEventsHandler(identitySvc))
}

func createNodeHandler(svc *services.IdentityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Name string `json:"name" validate:"required,max=255"`
		}
		if err := c.BodyParser(&req); err != nil {
			logging.Logger.Warn("Invalid request body", zap.Error(err))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
		}

		// Validate input (security first)
		if len(req.Name) > MaxNodeNameLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name too long"})
		}

		node, err := svc.CreateNode(req.Name)

		if err != nil {
			logging.Logger.Error("Failed to create node", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create node"})
		}

		metadata := map[string]interface{}{"node_id": node.ID}
		logging.Audit("Node created", metadata)

		return c.JSON(fiber.Map{"data": node})
	}
}

func generateTokenHandler(svc *services.IdentityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid node ID"})
		}

		token, err := svc.GenerateInstallToken(uint(id))
		if err != nil {
			logging.Logger.Error("Failed to generate token", zap.Error(err), zap.Int("node_id", id))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		return c.JSON(fiber.Map{"data": map[string]string{"token": token}})
	}
}

func revokeNodeHandler(svc *services.IdentityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid node ID"})
		}

		if err := svc.RevokeNode(uint(id)); err != nil {
			logging.Logger.Error("Failed to revoke node", zap.Error(err), zap.Int("node_id", id))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to revoke node"})
		}

		metadata := map[string]interface{}{"node_id": id}
		logging.Audit("Node revoked", metadata)

		return c.JSON(fiber.Map{"data": "Node revoked successfully"})
	}
}

func listNodesHandler(svc *services.IdentityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		nodes, err := svc.ListNodes()
		if err != nil {
			logging.Logger.Error("Failed to list nodes", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list nodes"})
		}
		return c.JSON(fiber.Map{"data": nodes})
	}
}

func getNodeHandler(svc *services.IdentityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid node ID"})
		}

		node, err := svc.GetNode(uint(id))
		if err != nil {
			logging.Logger.Error("Failed to get node", zap.Error(err), zap.Int("node_id", id))
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Node not found"})
		}

		return c.JSON(fiber.Map{"data": node})
	}
}

func getNodeMetricsHandler(svc *services.IdentityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid node ID"})
		}

		limit := c.QueryInt("limit", 100)
		if limit > 1000 {
			limit = 1000
		}

		metrics, err := svc.GetNodeMetrics(uint(id), limit)
		if err != nil {
			logging.Logger.Error("Failed to get node metrics", zap.Error(err), zap.Int("node_id", id))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get metrics"})
		}

		return c.JSON(fiber.Map{"data": metrics})
	}
}

func settingsHandler(svc *services.IdentityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		activeCount, err := svc.GetActiveNodeCount()
		if err != nil {
			logging.Logger.Warn("Failed to get active count", zap.Error(err))
			activeCount = 0
		}

		revokedCount, err := svc.GetRevokedNodeCount()
		if err != nil {
			logging.Logger.Warn("Failed to get revoked count", zap.Error(err))
			revokedCount = 0
		}

		alertCount, err := svc.GetTotalActiveAlerts()
		if err != nil {
			logging.Logger.Warn("Failed to get alert count", zap.Error(err))
			alertCount = 0
		}

		data := map[string]interface{}{
			"cert_health": map[string]int{
				"active":  activeCount,
				"revoked": revokedCount,
			},
			"alerts": map[string]int{
				"active": alertCount,
			},
		}

		return c.JSON(fiber.Map{"data": data})
	}
}

func listEventsHandler(svc *services.IdentityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit := c.QueryInt("limit", 100)
		if limit > 1000 {
			limit = 1000
		}

		events, err := svc.ListEvents(limit)
		if err != nil {
			logging.Logger.Error("Failed to list events", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list events"})
		}

		return c.JSON(fiber.Map{"data": events})
	}
}

// Task endpoints
func TasksRoutes(app *fiber.App, authMiddleware fiber.Handler, taskSvc *services.TaskService) {
	v1 := app.Group("/v1", authMiddleware)
	v1.Post("/tasks", func(c *fiber.Ctx) error {
		var req services.TaskRequest
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			logging.Logger.Warn("Invalid task request", zap.Error(err))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
		}
		task, replay, err := taskSvc.CreateTask(req)
		if err != nil {
			logging.Logger.Error("Failed to create task", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create task"})
		}
		meta := fiber.Map{"idempotent_replay": replay}
		return c.JSON(fiber.Map{"data": task, "meta": meta})
	})

	v1.Get("/tasks", func(c *fiber.Ctx) error {
		identity := c.Query("identity")
		tasks, err := taskSvc.ListTasks(identity, 100)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list"})
		}
		type tview struct {
			ID        uint      `json:"id"`
			Identity  string    `json:"identity"`
			Type      string    `json:"type"`
			Class     string    `json:"class"`
			Status    string    `json:"status"`
			CreatedAt time.Time `json:"created_at"`
		}
		rows := make([]tview, 0, len(tasks))
		for _, t := range tasks {
			rows = append(rows, tview{
				ID:        t.ID,
				Identity:  t.Identity,
				Type:      t.Type,
				Class:     t.Class,
				Status:    t.Status,
				CreatedAt: t.CreatedAt,
			})
		}
		return c.JSON(fiber.Map{"data": rows})
	})

	v1.Get("/tasks/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID"})
		}
		task, err := taskSvc.GetTaskByID(uint(id))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not found"})
		}
		return c.JSON(fiber.Map{"data": task})
	})
}
