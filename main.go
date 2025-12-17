package main

import (
    "log"

    "example.com/workflowapi/client"
    "example.com/workflowapi/handler"
    "example.com/workflowapi/model"

    "github.com/gin-gonic/gin"
)

func main() {
    db := client.InitDB()

    db.AutoMigrate(
        &model.Agent{},
        &model.Workflow{},
        &model.Step{},
        &model.N{},
    )

    r := gin.Default()

    handler.RegisterAgentRoutes(r, db)
    handler.RegisterWorkflowRoutes(r, db)
    handler.RegisterStepRoutes(r, db)
    handler.RegisterNRoutes(r, db)

    log.Fatal(r.Run(":8080"))
}
