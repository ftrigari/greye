package ports

import "github.com/gofiber/fiber/v2"

type ApiExposed interface {
	AddApplication(ctx *fiber.Ctx) error
	AddApplicationBySvc(ctx *fiber.Ctx) error
	UnscheduleApplication(ctx *fiber.Ctx) error
	MonitoringApplication(ctx *fiber.Ctx) error
	GetApplicationMonitored(ctx *fiber.Ctx) error
	GetApplicationMonitoredByPod(ctx *fiber.Ctx) error
	//Check(ctx *fiber.Ctx) error
}
