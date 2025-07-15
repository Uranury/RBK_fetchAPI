package models

type Summary struct {
	Response struct {
		Players []struct {
			SteamID                  string `json:"steamid"`
			CommunityVisibilityState int    `json:"communityvisibilitystate"`
			ProfileState             int    `json:"profilestate"`
			PersonaName              string `json:"personaname"`
			CommentPermission        int    `json:"commentpermission"`
			ProfileURL               string `json:"profileurl"`
			Avatar                   string `json:"avatar"`
			AvatarMedium             string `json:"avatarmedium"`
			AvatarFull               string `json:"avatarfull"`
			AvatarHash               string `json:"avatarhash"`
			LastLogoff               int    `json:"lastlogoff"`
			PersonaState             int    `json:"personastate"`
			RealName                 string `json:"realname"`
			PrimaryClanID            string `json:"primaryclanid"`
			TimeCreated              int    `json:"timecreated"`
			PersonaStateFlags        int    `json:"personastateflags"`
			LocCountryCode           string `json:"loccountrycode"`
			LocStateCode             string `json:"locstatecode"`
		} `json:"players"`
	} `json:"response"`
}
