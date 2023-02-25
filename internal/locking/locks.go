package locking

import "sync"

var (
	SpotifyClientLock sync.Mutex
	TwitchClientLock  sync.Mutex
)
