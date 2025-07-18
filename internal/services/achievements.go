package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Uranury/RBK_fetchAPI/internal/models"
)

const (
	fetchGameSchemaTemplate                   = "http://api.steampowered.com/ISteamUserStats/GetSchemaForGame/v2/?key=%s&appid=%s"
	fetchPlayerAchievementsTemplate           = "http://api.steampowered.com/ISteamUserStats/GetPlayerAchievements/v0001/?appid=%s&key=%s&steamid=%s"
	fetchGlobalAchievementPercentagesTemplate = "http://api.steampowered.com/ISteamUserStats/GetGlobalAchievementPercentagesForApp/v0002/?gameid=%s"
)

func (s *SteamService) GetPlayerAchievements(ctx context.Context, steamID, appID string) (*models.PlayerAchievements, error) {
	cacheKey := fmt.Sprintf("player_achievements:%s:game:%s", steamID, appID)

	cached, err := s.Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var achievements models.PlayerAchievements
		if err := json.Unmarshal([]byte(cached), &achievements); err == nil {
			return &achievements, nil
		}
	}

	// Get player achievements
	playerAchievements, err := s.fetchPlayerAchievements(ctx, steamID, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch player achievements: %w", err)
	}

	// Get game schema to get achievement names and descriptions
	gameSchema, err := s.fetchGameSchema(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch game schema: %w", err)
	}

	// Get global achievement percentages for rarity
	globalPercentages, err := s.fetchGlobalAchievementPercentages(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch global achievement percentages: %w", err)
	}

	// Create maps for quick lookup
	schemaMap := make(map[string]models.Achievement)
	for _, ach := range gameSchema.Game.AvailableGameStats.Achievements {
		schemaMap[ach.Name] = models.Achievement{
			Name:        ach.Name,
			DisplayName: ach.DisplayName,
			Description: ach.Description,
			Icon:        ach.Icon,
			IconGray:    ach.IconGray,
		}
	}

	percentageMap := make(map[string]float64)
	for _, perc := range globalPercentages.AchievementPercentages.Achievements {
		if percentage, err := strconv.ParseFloat(perc.Percent, 64); err == nil {
			percentageMap[perc.Name] = percentage
		}
	}

	// Combine player achievements with schema data and rarity
	result := &models.PlayerAchievements{
		SteamID:  playerAchievements.PlayerStats.SteamID,
		GameName: playerAchievements.PlayerStats.GameName,
	}

	for _, playerAch := range playerAchievements.PlayerStats.Achievements {
		if schemaAch, exists := schemaMap[playerAch.APIName]; exists {
			achievement := models.Achievement{
				Name:        schemaAch.Name,
				DisplayName: schemaAch.DisplayName,
				Description: schemaAch.Description,
				Achieved:    playerAch.Achieved == 1,
				Icon:        schemaAch.Icon,
				IconGray:    schemaAch.IconGray,
				Rarity:      percentageMap[playerAch.APIName], // Add rarity percentage
			}

			// Format unlock time correctly (convert from Unix timestamp)
			if playerAch.Achieved == 1 && playerAch.UnlockTime > 0 {
				achievement.UnlockTime = time.Unix(playerAch.UnlockTime, 0)
			}

			result.Achievements = append(result.Achievements, achievement)
		}
	}

	bytes, err := json.Marshal(result)
	if err == nil {
		s.Cache.Set(ctx, cacheKey, bytes, time.Minute*5)
	}

	return result, nil
}

func (s *SteamService) fetchPlayerAchievements(ctx context.Context, steamID, appID string) (*models.PlayerAchievementsResponse, error) {
	cacheKey := fmt.Sprintf("fetched_player_achievements:%s:game:%s", steamID, appID)

	cached, err := s.Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var achievements models.PlayerAchievementsResponse
		if err := json.Unmarshal([]byte(cached), &achievements); err == nil {
			return &achievements, nil
		}
	}

	url := fmt.Sprintf(fetchPlayerAchievementsTemplate, appID, s.APIKey, steamID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result models.PlayerAchievementsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.PlayerStats.Success {
		return nil, fmt.Errorf("steam API returned success=false")
	}

	bytes, err := json.Marshal(result)
	if err == nil {
		s.Cache.Set(ctx, cacheKey, bytes, time.Minute*5)
	}

	return &result, nil
}

func (s *SteamService) fetchGameSchema(ctx context.Context, appID string) (*models.GameSchemaResponse, error) {
	cacheKey := fmt.Sprintf("game_schema:%s", appID)

	cached, err := s.Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var schema models.GameSchemaResponse
		if err := json.Unmarshal([]byte(cached), &schema); err == nil {
			return &schema, nil
		}
	}

	url := fmt.Sprintf(fetchGameSchemaTemplate, s.APIKey, appID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result models.GameSchemaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(result)
	if err == nil {
		s.Cache.Set(ctx, cacheKey, bytes, time.Hour*336)
	}

	return &result, nil
}

func (s *SteamService) fetchGlobalAchievementPercentages(ctx context.Context, appID string) (*models.GlobalAchievementPercentagesResponse, error) {
	cacheKey := fmt.Sprintf("global_achievement_percentages:%s", appID)

	cached, err := s.Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var percentages models.GlobalAchievementPercentagesResponse
		if err := json.Unmarshal([]byte(cached), &percentages); err == nil {
			return &percentages, nil
		}
	}

	url := fmt.Sprintf(fetchGlobalAchievementPercentagesTemplate, appID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result models.GlobalAchievementPercentagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(result)
	if err == nil {
		// Cache global percentages for longer since they change slowly
		s.Cache.Set(ctx, cacheKey, bytes, time.Hour*24)
	}

	return &result, nil
}
