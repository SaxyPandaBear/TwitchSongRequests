package db

import (
	"fmt"

	"github.com/saxypandabear/twitchsongrequests/pkg/users"
)

type UserStore interface {
	GetUser(id string) (*users.User, error)
	AddUser(user *users.User) error
	DeleteUser(id string) error
}

// InMemoryUserStore is used for mocking and unit testing
type InMemoryUserStore struct {
	Data map[string]*users.User
}

var _ UserStore = (*InMemoryUserStore)(nil)

func (s *InMemoryUserStore) GetUser(id string) (*users.User, error) {
	user, ok := s.Data[id]
	if !ok {
		return nil, fmt.Errorf("user %s not found", id)
	}

	return user, nil
}

func (s *InMemoryUserStore) AddUser(user *users.User) error {
	s.Data[user.TwitchID] = user
	return nil // TODO: not sure if it's worth testing negative case
}

func (s *InMemoryUserStore) DeleteUser(id string) error {
	delete(s.Data, id)
	return nil
}
