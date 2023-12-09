package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
	"go.uber.org/zap"
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

func (s *PostgresUserStore) GetUser(id string) (*users.User, error) {
	u := users.User{
		TwitchID: id,
	}

	err := s.pool.QueryRow(context.Background(),
		"SELECT COALESCE(twitch_access, ''), COALESCE(twitch_refresh, ''), COALESCE(spotify_access, ''), COALESCE(spotify_refresh, ''), spotify_expiry, COALESCE(subscribed, FALSE), COALESCE(subscription_id, ''), COALESCE(email, '') FROM users WHERE id=$1", id).
		Scan(&u.TwitchAccessToken, &u.TwitchRefreshToken, &u.SpotifyAccessToken, &u.SpotifyRefreshToken, &u.SpotifyExpiry, &u.Subscribed, &u.SubscriptionID, &u.Email)

	if err != nil {
		zap.L().Error("failed to get user", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	return &u, nil
}

func (s *PostgresUserStore) AddUser(user *users.User) error {
	if _, err := s.pool.Exec(context.Background(),
		"INSERT INTO users(id, twitch_access, twitch_refresh, last_updated) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET twitch_access = $2, twitch_refresh = $3, last_updated = $4",
		user.TwitchID,
		user.TwitchAccessToken,
		user.TwitchRefreshToken,
		time.Now().Format(time.RFC3339)); err != nil {
		zap.L().Error("failed to insert user", zap.String("id", user.TwitchID), zap.Error(err))
		return err
	}
	return nil
}

func (s *PostgresUserStore) UpdateUser(user *users.User) error {
	if _, err := s.pool.Exec(context.Background(),
		"update users set twitch_access=$1, twitch_refresh=$2, spotify_access=$3, spotify_refresh=$4, spotify_expiry=$5, last_updated=$6, subscribed=$7, subscription_id=$8, email=$9 where id=$10",
		user.TwitchAccessToken,
		user.TwitchRefreshToken,
		user.SpotifyAccessToken,
		user.SpotifyRefreshToken,
		user.SpotifyExpiry,
		time.Now().Format(time.RFC3339),
		user.Subscribed,
		user.SubscriptionID,
		user.Email,
		user.TwitchID); err != nil {
		zap.L().Error("failed to update user", zap.String("id", user.TwitchID), zap.Error(err))
		return err
	}
	return nil
}

func (s *PostgresUserStore) DeleteUser(id string) error {
	if _, err := s.pool.Exec(context.Background(), "delete from users where id=$1", id); err != nil {
		zap.L().Error("failed to delete user", zap.String("id", id), zap.Error(err))
		return err
	}
	return nil
}
