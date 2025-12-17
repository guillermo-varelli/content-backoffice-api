package model

type Agent struct {
    ID       uint64 `gorm:"primaryKey"`
    Provider string
    Secret   string
}
