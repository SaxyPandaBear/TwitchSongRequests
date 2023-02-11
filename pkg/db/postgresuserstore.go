package db

import (
	"github.com/jackc/pgx/v5"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
)

var _ UserStore = (*PostgresUserStore)(nil)

type PostgresUserStore struct {
	conn *pgx.Conn
}

func NewPostgresUserStore(c *pgx.Conn) *PostgresUserStore {
	return &PostgresUserStore{
		conn: c,
	}
}

func (db *PostgresUserStore) GetUser(id string) (*users.User, error) {
	return nil, nil
}

func (db *PostgresUserStore) AddUser(user *users.User) error {
	return nil
}

func (db *PostgresUserStore) DeleteUser(id string) error {
	return nil
}
