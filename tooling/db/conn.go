package db

import (
	"database/sql"
	"github.com/rusinikita/system-design-trainer/med-care-app-cache/db"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func Conn() *sql.DB {
	// Load .env file from the root directory
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := sql.Open(os.Getenv("DB_DRIVER"), os.Getenv("DB_CONNECT"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db
}

func RepoConn() (*sql.DB, *db.DashboardRepository) {
	d := Conn()

	return d, db.NewDashboardRepository(d)
}
