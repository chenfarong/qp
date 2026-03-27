package db

import (
	"testing"

	"github.com/aoyo/qp/pkg/db"
)

func TestInitDB(t *testing.T) {
	uri := "mongodb://admin:password@localhost:27017/qp_game?authSource=admin"

	dbInstance, err := db.InitDB(uri)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbInstance.Close()

	t.Log("Database connection successful!")
}
