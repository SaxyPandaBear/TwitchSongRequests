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

func (p *PostgresMessageCounter) TotalMessages() uint64 {
	var v uint64
	if err := p.pool.QueryRow(context.Background(), "select count(*) from messages").Scan(&v); err != nil {
		log.Println("failed to count messages", err)
	}
	return v
}

func (p *PostgresMessageCounter) RunningCount(days int) uint64 {
	var v uint64
	if err := p.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM messages WHERE created_at > now() - interval '$1' day", days).Scan(v); err != nil {
		log.Println("failed to get running count of messages", err)
	}
	return v
}
