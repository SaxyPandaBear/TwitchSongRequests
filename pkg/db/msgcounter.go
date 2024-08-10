package db

import (
	"github.com/saxypandabear/twitchsongrequests/pkg/o11y/metrics"
)

type MessageCounter interface {
	AddMessage(*metrics.Message)
	TotalMessages() uint64
	RunningCount(int) uint64
	MessagesForUser(string) []*metrics.Message
}

type NoopMessageCounter struct{}

// AddMessage implements MessageCounter.
func (n *NoopMessageCounter) AddMessage(*metrics.Message) {}

// MessagesForUser implements MessageCounter.
func (n *NoopMessageCounter) MessagesForUser(string) []*metrics.Message {
	return nil
}

// RunningCount implements MessageCounter.
func (n *NoopMessageCounter) RunningCount(int) uint64 {
	return 0
}

// TotalMessages implements MessageCounter.
func (n *NoopMessageCounter) TotalMessages() uint64 {
	return 0
}

var _ MessageCounter = (*NoopMessageCounter)(nil)
