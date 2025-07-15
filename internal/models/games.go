package models

type OwnedGamesResponse struct {
	Response struct {
		GameCount int `json:"game_count"`
		Games     []struct {
			AppID                    int    `json:"appid"`
			Name                     string `json:"name"`
			PlaytimeForever          int    `json:"playtime_forever"`
			ImgIconURL               string `json:"img_icon_url"`
			ImgLogoURL               string `json:"img_logo_url"`
			HasCommunityVisibleStats bool   `json:"has_community_visible_stats,omitempty"`
		} `json:"games"`
	} `json:"response"`
}

type AchievementInfo struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	Hidden      bool    `json:"hidden"`
	Rarity      float64 `json:"rarity"`
}
