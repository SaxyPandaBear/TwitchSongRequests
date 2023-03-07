package preferences

type Preference struct {
	TwitchID       string `column:"id"`
	ExplicitSongs  bool   `column:"explicit"`
	CustomRewardID string `column:"reward_id"`
}
