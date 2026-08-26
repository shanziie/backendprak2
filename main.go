package main

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// middleware untuk ngecek Content-Type harus application/json. kalo ga sesuai bakal return 415 Unsupported Media Type
func requireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json") // 415[cite: 1]
		}
	}
	return c.Next()
}

func main() {
	app := fiber.New()

	app.Use(requestid.New())
	app.Use(logger.New())

	api := app.Group("/api/v1")
	
	// pasang requireJSON middleware di semua route /api/v1/students
	studentsRoute := api.Group("/students", requireJSON)
	studentsRoute.Get("/", listStudents)
	studentsRoute.Get("/:id", getStudent)
	studentsRoute.Post("/", createStudent)
	studentsRoute.Put("/:id", replaceStudent)
	studentsRoute.Patch("/:id", patchStudent)
	studentsRoute.Delete("/:id", deleteStudent)

	app.Use(func(c *fiber.Ctx) error {
		return fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	log.Println("Server jalan di port 3000")
	log.Fatal(app.Listen(":3000"))
}