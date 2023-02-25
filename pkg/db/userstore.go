package db

import (
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
)

type UserStore interface {
	GetUser(id string) (*users.User, error)
	AddUser(user *users.User) error
	UpdateUser(user *users.User) error
	DeleteUser(id string) error
}
