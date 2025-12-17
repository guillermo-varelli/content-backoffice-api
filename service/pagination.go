package service

import "gorm.io/gorm"

func Paginate(page, size int) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        if page < 1 {
            page = 1
        }
        if size < 1 {
            size = 10
        }
        offset := (page - 1) * size
        return db.Offset(offset).Limit(size)
    }
}
