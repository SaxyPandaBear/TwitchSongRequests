package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
)

var _ UserStore = (*PostgresUserStore)(nil)

type PostgresUserStore struct {
	pool *pgxpool.Pool
}

func NewPostgresUserStore(pool *pgxpool.Pool) *PostgresUserStore {
	return &PostgresUserStore{
		pool: pool,
	}
}

func (db *PostgresUserStore) GetUser(id string) (*users.User, error) {
	u := users.User{
		TwitchID: id,
	}

	err := db.pool.QueryRow(context.Background(),
		"select twitch_access, twitch_refresh, spotify_access, spotify_refresh from users where id=$1", id).
		Scan(&u.TwitchAccessToken, &u.TwitchRefreshToken, &u.SpotifyAccessToken, &u.SpotifyRefreshToken)

	if err != nil {
		log.Printf("failed to get user %s: %v\n", id, err)
		return nil, err
	}
	return &u, nil
}

func (db *PostgresUserStore) AddUser(user *users.User) error {
	if _, err := db.pool.Exec(context.Background(),
		"insert into users(id, twitch_access, twitch_refresh, last_updated) values ($1, $2, $3, $4) on conflict do nothing",
		user.TwitchID,
		user.TwitchAccessToken,
		user.TwitchRefreshToken,
		time.Now().Format(time.RFC3339)); err != nil {
		log.Printf("failed to insert user %s: %v\n", user.TwitchID, err)
		return err
	}
	return nil
}

func (db *PostgresUserStore) UpdateUser(user *users.User) error {
	return nil // TODO: this needs to handle Twitch and Spotify. Maybe need to split table?
}

func (db *PostgresUserStore) DeleteUser(id string) error {
	if _, err := db.pool.Exec(context.Background(), "delete from users where id=$1", id); err != nil {
		log.Printf("failed to delete user %s: %v\n", id, err)
		return err
	}
	return nil
}
