package testutil

import (
	"context"
	"errors"
	"fmt"

	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
	"github.com/zmb3/spotify/v2"
)

type DummyPublisher struct {
	Messages   chan string
	ShouldFail bool
}

var _ queue.Publisher = (*DummyPublisher)(nil)

func (p DummyPublisher) Publish(client queue.Queuer, url string) error {
	if p.ShouldFail {
		return errors.New("oops")
	}

	p.Messages <- url
	return nil
}

type MockReadCloser struct{}

func (m MockReadCloser) Read(p []byte) (int, error) {
	return 0, errors.New("expected to fail")
}
func (m MockReadCloser) Close() error {
	return nil
}

type MockQueuer struct {
	ShouldFail bool
	Messages   []spotify.ID
}

func (m *MockQueuer) QueueSong(ctx context.Context, trackID spotify.ID) error {
	if m.ShouldFail {
		return errors.New("expected to fail")
	}

	m.Messages = append(m.Messages, trackID)
	return nil
}

// InMemoryUserStore is used for mocking and unit testing
type InMemoryUserStore struct {
	Data map[string]*users.User
}

var _ db.UserStore = (*InMemoryUserStore)(nil)

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

func (s *InMemoryUserStore) UpdateUser(user *users.User) error {
	s.Data[user.TwitchID] = user
	return nil
}

func (s *InMemoryUserStore) DeleteUser(id string) error {
	delete(s.Data, id)
	return nil
}
