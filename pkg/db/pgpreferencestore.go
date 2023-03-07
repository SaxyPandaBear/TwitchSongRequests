package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saxypandabear/twitchsongrequests/pkg/preferences"
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

	err := s.pool.QueryRow(context.Background(), "select COALESCE(explicit, false), COALESCE(reward_id, '') from preferences where id=$1", id).
		Scan(&p.ExplicitSongs, &p.CustomRewardID)
	if err != nil {
		log.Printf("failed to get user %s: %v\n", id, err)
		return nil, err
	}

	return &p, nil
}

func (s *PostgresPreferenceStore) AddPreference(p *preferences.Preference) error {
	if _, err := s.pool.Exec(context.Background(),
		"insert into preferences(id, reward_id, explicit, last_updated) values ($1, $2, $3, $4) on conflict do nothing",
		p.TwitchID,
		p.CustomRewardID,
		p.ExplicitSongs,
		time.Now().Format(time.RFC3339)); err != nil {
		log.Printf("failed to insert preferences for %s: %v\n", p.TwitchID, err)
		return err
	}
	return nil
}

func (s *PostgresPreferenceStore) UpdatePreference(p *preferences.Preference) error {
	if _, err := s.pool.Exec(context.Background(),
		"update preferences set reward_id=$1, explicit=$2, last_updated=$3 where id=$4",
		p.CustomRewardID,
		p.ExplicitSongs,
		time.Now().Format(time.RFC3339),
		p.TwitchID); err != nil {
		log.Printf("failed to update preferences for %s: %v\n", p.TwitchID, err)
		return err
	}
	return nil
}

func (s *PostgresPreferenceStore) DeletePreference(id string) error {
	if _, err := s.pool.Exec(context.Background(), "delete from preferences where id=$1", id); err != nil {
		log.Printf("failed to delete user %s: %v\n", id, err)
		return err
	}

	return nil
}
