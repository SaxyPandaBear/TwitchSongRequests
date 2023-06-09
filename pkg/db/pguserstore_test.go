package db_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/stretchr/testify/assert"
)

var connectOnce sync.Once
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

func TestPostgresGetUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	connectOnce.Do(connect)

	store := db.NewPostgresUserStore(pool)
	u, err := store.GetUser("12345")
	assert.NoError(t, err)
	assert.NotNil(t, u)
}
