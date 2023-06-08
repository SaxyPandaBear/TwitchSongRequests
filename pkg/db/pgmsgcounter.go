package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saxypandabear/twitchsongrequests/pkg/o11y/metrics"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const LIMIT = 100

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
	if _, err := p.pool.Exec(context.Background(), "insert into messages(created_at, success, broadcaster_id, spotify_track) values ($1, $2, $3, $4)", m.CreatedAt, m.Success, m.BroadcasterID, m.SpotifyTrack); err != nil {
		zap.L().Error("failed to add message", zap.Error(err))
	}
}

func (p *PostgresMessageCounter) TotalMessages() uint64 {
	var v uint64
	if err := p.pool.QueryRow(context.Background(), "select count(id) from messages where success = 1").Scan(&v); err != nil {
		zap.L().Error("failed to count messages", zap.Error(err))
	}
	return v
}

func (p *PostgresMessageCounter) RunningCount(days int) uint64 {
	var v uint64
	// https://github.com/jackc/pgx/issues/852 can't embed the parameter directly in the text string for the interval syntax
	if err := p.pool.QueryRow(context.Background(), "SELECT COUNT(id) FROM messages WHERE success = 1 AND AGE(messages.created_at) <= $1 * INTERVAL '1 day'", days).Scan(&v); err != nil {
		zap.L().Error("failed to get running count of messages", zap.Error(err))
	}
	return v
}

func (p *PostgresMessageCounter) MessagesForUser(id string) []*metrics.Message {
	rows, err := p.pool.Query(context.Background(), "SELECT spotify_track, success FROM messages WHERE broadcaster_id = $1 AND spotify_track != '' ORDER BY id DESC LIMIT 100", id)
	if err != nil {
		zap.L().Error("failed to query for messages", zap.Error(err))
		return []*metrics.Message{}
	}
	m := make([]*metrics.Message, 0, LIMIT)
	var multi error
	for rows.Next() {
		var msg metrics.Message
		if err = rows.Scan(&msg); err != nil {
			multi = multierr.Append(multi, err)
		} else {
			m = append(m, &msg)
		}
	}
	if multi != nil {
		zap.L().Error("errors occurred while scanning messages", zap.Error(multi))
	}

	return m
}
