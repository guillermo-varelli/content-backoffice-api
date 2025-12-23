package main

import (
	"log"

	"example.com/workflowapi/client"
	"example.com/workflowapi/config"
	"example.com/workflowapi/handler"
	"example.com/workflowapi/model"

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
		&model.N{},
	)

	r := gin.Default()

	// Rutas de autenticación (públicas)
	handler.RegisterAuthRoutes(r, db, cfg)

	// Rutas de gestión de usuarios (requieren autenticación y scope users:admin)
	handler.RegisterUserRoutes(r, db, cfg)

	// Rutas protegidas con JWT y scopes
	handler.RegisterAgentRoutes(r, db, cfg)
	handler.RegisterWorkflowRoutes(r, db, cfg)
	handler.RegisterStepRoutes(r, db, cfg)
	handler.RegisterStepExecutionRoutes(r, db, cfg)
	handler.RegisterNRoutes(r, db, cfg)

	log.Fatal(r.Run(":8080"))
}
