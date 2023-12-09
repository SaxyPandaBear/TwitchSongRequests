package constants

const (
	// Auth keys
	TwitchClientIDKey       = "TWITCH_CLIENT_ID"
	TwitchClientSecretKey   = "TWITCH_CLIENT_SECRET" //nolint: gosec
	TwitchEventSubSecretKey = "TWITCH_SECRET"
	SpotifyClientIDKey      = "SPOTIFY_CLIENT_ID"
	SpotifyClientSecretKey  = "SPOTIFY_CLIENT_SECRET" //nolint: gosec

	// Server configs
	MockServerURLKey   = "MOCK_SERVER_URL"
	TwitchRedirectURL  = "TWITCH_REDIRECT_URL"
	SpotifyRedirectURL = "SPOTIFY_REDIRECT_URL"
	SiteRedirectURL    = "SITE_REDIRECT_URL"
	SpotifyStateKey    = "SPOTIFY_STATE"
	TwitchStateKey     = "TWITCH_STATE"

	// Shared cookie
	TwitchIDCookieKey = "TwitchSongRequests-Twitch-ID"

	// User metrics
	NumOnboardedUsers = "ONBOARDED_USERS"
	NumAllowedUsers   = "ALLOWED_USERS"
)
