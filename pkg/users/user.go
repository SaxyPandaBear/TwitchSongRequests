package users

type User struct {
	TwitchID string
	TwitchAuth
}

type TwitchAuth struct {
	AccessToken  string
	RefreshToken string
}

type SpotifyAuth struct {
	AccessToken  string
	RefreshToken string
}
