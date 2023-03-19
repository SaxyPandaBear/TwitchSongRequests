package db

import (
	"github.com/saxypandabear/twitchsongrequests/pkg/o11y/metrics"
)

type MessageCounter interface {
	AddMessage(*metrics.Message)
}
