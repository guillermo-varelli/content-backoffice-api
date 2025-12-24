package service

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

func ApplyStepExecutionFilters(db *gorm.DB, c *gin.Context) (*gorm.DB, error) {

	// id
	if v := c.Query("id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, err
		}
		db = db.Where("step_executions.id = ?", id)
	}

	// execution_id
	if v := c.Query("execution_id"); v == "" {
		v = c.Query("executionId")
	} else if v != "" {
		execID, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, err
		}
		db = db.Where("step_executions.execution_id = ?", execID)
	}

	// workflow_id (JOIN MANUAL, NO preload)
	if v := c.Query("workflow_id"); v == "" {
		v = c.Query("workflowId")
	} else if v != "" {
		wfID, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, err
		}
		db = db.
			Joins("JOIN executions ON executions.id = step_executions.execution_id").
			Where("executions.workflow_id = ?", wfID)
	}

	return db, nil
}
