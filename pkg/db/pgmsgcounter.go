package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saxypandabear/twitchsongrequests/pkg/o11y/metrics"
)

var _ MessageCounter = (*PostgresMessageCounter)(nil)

func NewPostgresMessageCounter(pool *pgxpool.Pool) *PostgresMessageCounter {
	return &PostgresMessageCounter{
		pool: pool,
	}
}

type PostgresMessageCounter struct {
	pool *pgxpool.Pool
}

func (p *PostgresMessageCounter) AddMessage(m *metrics.Message) {
	if _, err := p.pool.Exec(context.Background(), "insert into messages(created_at, success) values ($1, $2)", m.CreatedAt, m.Success); err != nil {
		log.Println("failed to add message", err)
	}
}

func (p *PostgresMessageCounter) CountMessages() uint64 {
	var v uint64
	err := p.pool.QueryRow(context.Background(), "select count(*) from messages").Scan(&v)

	if err != nil {
		log.Println("failed to count messages", err)
	}
	return v
}
