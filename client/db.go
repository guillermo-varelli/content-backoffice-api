package client

import (
    "fmt"
    "log"

    "example.com/workflowapi/config"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func InitDB() *gorm.DB {
    cfg := config.Load()

    dsn := fmt.Sprintf(
        "%s:%s@tcp(%s:%s)/%s?parseTime=true",
        cfg.DBUser,
        cfg.DBPass,
        cfg.DBHost,
        cfg.DBPort,
        cfg.DBName,
    )

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    return db
}
