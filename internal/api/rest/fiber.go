package rest

import (
	"github.com/gofiber/fiber/v2"
)

func InitFiber() *fiber.App {
	return fiber.New()
}

func HandlerFiber() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello from fiber rest!")
	}
}
