package main

import (
	"log"
	"time"

	"example.com/workflowapi/client"
	"example.com/workflowapi/config"
	"example.com/workflowapi/handler"
	"example.com/workflowapi/model"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	db := client.InitDB()

	db.AutoMigrate(
		&model.User{},
		&model.Agent{},
		&model.Workflow{},
		&model.Step{},
		&model.Execution{},
		&model.StepExecution{},
		&model.Content{},
	)

	r := gin.Default()

	// ✅ CORS CONFIG (ESTO ES CLAVE)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // en prod poné tu dominio
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// ---------------- ROUTES ----------------

	// Públicas
	handler.RegisterAuthRoutes(r, db, cfg)

	// Protegidas
	handler.RegisterUserRoutes(r, db, cfg)
	handler.RegisterAgentRoutes(r, db, cfg)
	handler.RegisterWorkflowRoutes(r, db, cfg)
	handler.RegisterStepRoutes(r, db, cfg)
	handler.RegisterStepExecutionRoutes(r, db, cfg)
	handler.RegisterContentReviewRoutes(r, db, cfg)

	log.Fatal(r.Run(":8080"))
}
