package db

import (
	"context"
	"log"

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

	err := s.pool.QueryRow(context.Background(), "").Scan()
	if err != nil {
		log.Printf("failed to get user %s: %v\n", id, err)
		return nil, err
	}

	return &p, nil
}

func (s *PostgresPreferenceStore) AddPreference(*preferences.Preference) error {
	// TODO: implement
	return nil
}

func (s *PostgresPreferenceStore) UpdatePreference(*preferences.Preference) error {
	// TODO: implement
	return nil
}
func (s *PostgresPreferenceStore) DeletePreference(id string) error {
	if _, err := s.pool.Exec(context.Background(), "delete from preferences where id=$1", id); err != nil {
		log.Printf("failed to delete user %s: %v\n", id, err)
		return err
	}

	return nil
}
