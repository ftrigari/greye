package ports

import "github.com/gofiber/fiber/v2"

type HelloHandlers interface {
	Salutfa(ctx *fiber.Ctx) error
	SalutaDfiPiu(ctx *fiber.Ctx) error
}
