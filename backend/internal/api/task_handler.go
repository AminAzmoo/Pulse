package api

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"pulse/backend/internal/services"
	"pulse/backend/pkg/external/logging"
)

func createNodeTaskHandler(identitySvc *services.IdentityService, taskSvc *services.TaskService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid node ID"})
		}

		// Look up node to get Identity
		node, err := identitySvc.GetNode(uint(id))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Node not found"})
		}

		var req struct {
			Type   string                 `json:"type" validate:"required"`
			Class  string                 `json:"class" validate:"required"`
			Params map[string]interface{} `json:"params"`
		}

		if err := c.BodyParser(&req); err != nil {
			logging.Logger.Warn("Invalid task request body", zap.Error(err))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
		}

		taskReq := services.TaskRequest{
			Identity: node.Identity,
			Type:     req.Type,
			Class:    req.Class,
			Params:   req.Params,
		}

		task, replay, err := taskSvc.CreateTask(taskReq)
		if err != nil {
			logging.Logger.Error("Failed to create task", zap.Error(err))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		meta := fiber.Map{"idempotent_replay": replay}
		return c.JSON(fiber.Map{"data": task, "meta": meta})
	}
}
