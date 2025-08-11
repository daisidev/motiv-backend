
package middleware

import (
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/hidenkeys/motiv-backend/models"
)

func AuthRequired(jwtSecret []byte) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtSecret,
	})
}

func RoleRequired(roles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userRole := models.UserRole(claims["role"].(string))

		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}
}
