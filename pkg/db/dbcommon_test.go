package db_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func connect() {
	user := os.Getenv("POSTGRES_USER")
	pwd := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	var err error
	pool, err = pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/testdb", user, pwd, host, port))
	if err != nil {
		log.Fatalf("failed to connect to postgres db: %v\n", err)
	}
}
