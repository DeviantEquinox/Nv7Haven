package nv7haven

import (
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (d *Nv7Haven) getIP(c *fiber.Ctx) error {
	return c.SendString(net.ParseIP(strings.Split(c.Get("X-Forwarded-For"), ",")[0]).String())
}