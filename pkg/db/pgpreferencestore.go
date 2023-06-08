package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saxypandabear/twitchsongrequests/pkg/preferences"
	"go.uber.org/zap"
)

var _ PreferenceStore = (*PostgresPreferenceStore)(nil)

type PostgresPreferenceStore struct {
	pool *pgxpool.Pool
}

func NewPostgresPreferenceStore(pool *pgxpool.Pool) *PostgresPreferenceStore {
	return &PostgresPreferenceStore{
		pool: pool,
	}
}

func (s *PostgresPreferenceStore) GetPreference(id string) (*preferences.Preference, error) {
	p := preferences.Preference{
		TwitchID: id,
	}

	err := s.pool.QueryRow(context.Background(), "select COALESCE(explicit, false), COALESCE(reward_id, ''), COALESCE(max_song_length, 0) from preferences where id=$1", id).
		Scan(&p.ExplicitSongs, &p.CustomRewardID, &p.MaxSongLength)
	if err != nil {
		zap.L().Error("failed to get user", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	return &p, nil
}

func (s *PostgresPreferenceStore) AddPreference(p *preferences.Preference) error {
	if _, err := s.pool.Exec(context.Background(),
		"insert into preferences(id, reward_id, explicit, max_song_length, last_updated) values ($1, $2, $3, $4, $5) on conflict do nothing",
		p.TwitchID,
		p.CustomRewardID,
		p.ExplicitSongs,
		p.MaxSongLength,
		time.Now().Format(time.RFC3339)); err != nil {
		zap.L().Error("failed to insert preferences", zap.String("id", p.TwitchID), zap.Error(err))
		return err
	}
	return nil
}

func (s *PostgresPreferenceStore) UpdatePreference(p *preferences.Preference) error {
	if _, err := s.pool.Exec(context.Background(),
		"update preferences set reward_id=$1, explicit=$2, max_song_length=$3, last_updated=$4 where id=$5",
		p.CustomRewardID,
		p.ExplicitSongs,
		p.MaxSongLength,
		time.Now().Format(time.RFC3339),
		p.TwitchID); err != nil {
		zap.L().Error("failed to update preferences", zap.String("id", p.TwitchID), zap.Error(err))
		return err
	}
	return nil
}

func (s *PostgresPreferenceStore) DeletePreference(id string) error {
	if _, err := s.pool.Exec(context.Background(), "delete from preferences where id=$1", id); err != nil {
		zap.L().Error("failed to delete user", zap.String("id", id), zap.Error(err))
		return err
	}

	return nil
}
