package twitch

import "time"

type BearerToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type User struct {
	Data []struct {
		ID              string    `json:"id"`
		Login           string    `json:"login"`
		DisplayName     string    `json:"display_name"`
		Type            string    `json:"type"`
		BroadcasterType string    `json:"broadcaster_type"`
		Description     string    `json:"description"`
		ProfileImageURL string    `json:"profile_image_url"`
		OfflineImageURL string    `json:"offline_image_url"`
		ViewCount       int       `json:"view_count"`
		Email           string    `json:"email"`
		CreatedAt       time.Time `json:"created_at"`
	} `json:"data"`
}

type BroadcastProperties struct {
	GameID              *string `json:"game_id,omitempty"`
	BroadcasterLanguage *string `json:"broadcaster_language,omitempty"`
	Title               *string `json:"title,omitempty"`
	Delay               *int    `json:"delay,omitempty"`
}

type UserAuth struct {
	Client_id  string   `json:"client_id"`
	Login      string   `json:"login"`
	Scopes     []string `json:"scopes"`
	User_id    string   `json:"user_id"`
	Expires_in int      `json:"expires_in"`
}

type ChannelInfo struct {
	Data []struct {
		BroadcasterID               string   `json:"broadcaster_id"`
		BroadcasterLogin            string   `json:"broadcaster_login"`
		BroadcasterName             string   `json:"broadcaster_name"`
		BroadcasterLanguage         string   `json:"broadcaster_language"`
		GameID                      string   `json:"game_id"`
		GameName                    string   `json:"game_name"`
		Title                       string   `json:"title"`
		Delay                       int      `json:"delay"`
		Tags                        []string `json:"tags"`
		ContentClassificationLabels []string `json:"content_classification_labels"`
		IsBrandedContent            bool     `json:"is_branded_content"`
	} `json:"data"`
}
