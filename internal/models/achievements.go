package models

import "time"

type PlayerAchievementsResponse struct {
	PlayerStats struct {
		SteamID      string `json:"steamID"`
		GameName     string `json:"gameName"`
		Achievements []struct {
			APIName    string `json:"apiname"`
			Achieved   int    `json:"achieved"`
			UnlockTime int64  `json:"unlocktime"`
		} `json:"achievements"`
		Success bool `json:"success"`
	} `json:"playerstats"`
}

type GameSchemaResponse struct {
	Game struct {
		GameName           string `json:"gameName"`
		GameVersion        string `json:"gameVersion"`
		AvailableGameStats struct {
			Achievements []struct {
				Name         string `json:"name"`
				DefaultValue int    `json:"defaultvalue"`
				DisplayName  string `json:"displayName"`
				Hidden       int    `json:"hidden"`
				Description  string `json:"description"`
				Icon         string `json:"icon"`
				IconGray     string `json:"icongray"`
			} `json:"achievements"`
		} `json:"availableGameStats"`
	} `json:"game"`
}

type GlobalAchievementPercentagesResponse struct {
	AchievementPercentages struct {
		Achievements []struct {
			Name    string `json:"name"`
			Percent string `json:"percent"`
		} `json:"achievements"`
	} `json:"achievementpercentages"`
}

// Final processed achievement model
type Achievement struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
	Achieved    bool      `json:"achieved"`
	UnlockTime  time.Time `json:"unlockTime,omitempty"`
	Icon        string    `json:"icon"`
	IconGray    string    `json:"iconGray"`
	Rarity      float64   `json:"rarity"` // Percentage of players who have this achievement
}

type PlayerAchievements struct {
	SteamID      string        `json:"steamID"`
	GameName     string        `json:"gameName"`
	Achievements []Achievement `json:"achievements"`
}
