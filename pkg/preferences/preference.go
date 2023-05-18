package preferences

type Preference struct {
	TwitchID       string `column:"id"`
	ExplicitSongs  bool   `column:"explicit"`
	CustomRewardID string `column:"reward_id"`
	MaxSongLength  int    `column:"max_song_length" unit:"milliseconds"`
}
