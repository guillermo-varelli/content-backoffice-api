package service

import (
	"testing"

	"example.com/workflowapi/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPaginateAppliesLimitAndOffset(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("opening in-memory sqlite: %v", err)
	}

	stmt := db.Session(&gorm.Session{DryRun: true}).Scopes(Paginate(2, 5)).Find(&[]model.Agent{})
	if stmt.Error != nil {
		t.Fatalf("running dry-run query: %v", stmt.Error)
	}

	if stmt.Statement.Offset != 5 {
		t.Fatalf("expected offset 5, got %d", stmt.Statement.Offset)
	}

	if stmt.Statement.Limit == nil || *stmt.Statement.Limit != 5 {
		t.Fatalf("expected limit 5, got %v", stmt.Statement.Limit)
	}
}

func TestPaginateDefaultsWhenInputIsInvalid(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("opening in-memory sqlite: %v", err)
	}

	stmt := db.Session(&gorm.Session{DryRun: true}).Scopes(Paginate(0, 0)).Find(&[]model.Agent{})
	if stmt.Error != nil {
		t.Fatalf("running dry-run query: %v", stmt.Error)
	}

	if stmt.Statement.Offset != 0 {
		t.Fatalf("expected offset 0, got %d", stmt.Statement.Offset)
	}

	if stmt.Statement.Limit == nil || *stmt.Statement.Limit != 10 {
		t.Fatalf("expected default limit 10, got %v", stmt.Statement.Limit)
	}
}
