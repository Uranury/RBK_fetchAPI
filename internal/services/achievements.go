package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Uranury/RBK_fetchAPI/internal/apperrors"
	"github.com/Uranury/RBK_fetchAPI/internal/models"
)

const (
	fetchGameSchemaTemplate                   = "http://api.steampowered.com/ISteamUserStats/GetSchemaForGame/v2/?key=%s&appid=%s"
	fetchPlayerAchievementsTemplate           = "http://api.steampowered.com/ISteamUserStats/GetPlayerAchievements/v0001/?appid=%s&key=%s&steamid=%s"
	fetchGlobalAchievementPercentagesTemplate = "http://api.steampowered.com/ISteamUserStats/GetGlobalAchievementPercentagesForApp/v0002/?gameid=%s"
)

func (s *SteamService) GetPlayerAchievements(ctx context.Context, steamID, appID string) (*models.PlayerAchievements, *apperrors.APIError) {
	cacheKey := fmt.Sprintf("player_achievements:%s:game:%s", steamID, appID)

	cached, err := s.Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var achievements models.PlayerAchievements
		if err := json.Unmarshal([]byte(cached), &achievements); err == nil {
			return &achievements, nil
		}
	}

	playerAchievements, apiError := s.fetchPlayerAchievements(ctx, steamID, appID)
	if apiError != nil {
		return nil, apiError
	}

	gameSchema, apiError := s.fetchGameSchema(ctx, appID)
	if apiError != nil {
		return nil, apiError
	}

	// Get global achievement percentages for rarity
	globalPercentages, err := s.fetchGlobalAchievementPercentages(ctx, appID)
	if err != nil {
		return nil, apiError
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
		if err := s.Cache.Set(ctx, cacheKey, bytes, time.Minute*5).Err(); err != nil {
			log.Printf("failed to cache GetAchievements: %v", err)
		}
	}

	return result, nil
}

func (s *SteamService) fetchPlayerAchievements(ctx context.Context, steamID, appID string) (*models.PlayerAchievementsResponse, *apperrors.APIError) {
	cacheKey := fmt.Sprintf("fetched_player_achievements:%s:game:%s", steamID, appID)

	cached, err := s.Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var achievements models.PlayerAchievementsResponse
		if err := json.Unmarshal([]byte(cached), &achievements); err == nil {
			return &achievements, nil
		}
		log.Printf("failed to unmarshal cached Achievements: %v", err)
	}

	url := fmt.Sprintf(fetchPlayerAchievementsTemplate, appID, s.APIKey, steamID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, apperrors.WrapAPIError(500, err, "fetchPlayerAchievements request creation failed")
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, apperrors.WrapAPIError(500, err, "fetchPlayerAchievements request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("steam API responded with status: %s", resp.Status)
		return nil, apperrors.WrapAPIError(resp.StatusCode, err, "")
	}

	var result models.PlayerAchievementsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, apperrors.WrapAPIError(500, err, "fetchPlayerAchievements failed to decode JSON")
	}

	if !result.PlayerStats.Success {
		return nil, apperrors.WrapAPIError(409, err, "fetchPlayerAchievements, invalid appID or private profile")
	}

	bytes, err := json.Marshal(result)
	if err == nil {
		if err := s.Cache.Set(ctx, cacheKey, bytes, 5*time.Minute).Err(); err != nil {
			log.Printf("failed to cache fetchedAchievements: %v", err)
		}
	}

	return &result, nil
}

func (s *SteamService) fetchGameSchema(ctx context.Context, appID string) (*models.GameSchemaResponse, *apperrors.APIError) {
	cacheKey := fmt.Sprintf("game_schema:%s", appID)

	cached, err := s.Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var schema models.GameSchemaResponse
		if err := json.Unmarshal([]byte(cached), &schema); err == nil {
			return &schema, nil
		}
		log.Printf("Failed to get cached GameSchema: %v", err)
	}

	url := fmt.Sprintf(fetchGameSchemaTemplate, s.APIKey, appID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, apperrors.WrapAPIError(500, err, "fetchGameSchema request creation failed")
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, apperrors.WrapAPIError(500, err, "fetchGameSchema request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("API request failed with status: %d", resp.StatusCode)
		return nil, apperrors.WrapAPIError(resp.StatusCode, err, "")
	}

	var result models.GameSchemaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, apperrors.WrapAPIError(500, err, "fetchGameSchema failed to decode JSON")
	}

	bytes, err := json.Marshal(result)
	if err == nil {
		if err := s.Cache.Set(ctx, cacheKey, bytes, time.Hour*336).Err(); err != nil {
			log.Printf("failed to cache fetchedGameSchemas: %v", err)
		}
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
		log.Printf("failed to get cached achievement percentages: %v", err)
	}

	url := fmt.Sprintf(fetchGlobalAchievementPercentagesTemplate, appID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, apperrors.WrapAPIError(500, err, "fetchGlobalAchievementPercentages request creation failed")
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, apperrors.WrapAPIError(500, err, "fetchGlobalAchievementPercentages request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("API request failed with status: %d", resp.StatusCode)
		return nil, apperrors.WrapAPIError(resp.StatusCode, err, "")
	}

	var result models.GlobalAchievementPercentagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, apperrors.WrapAPIError(500, err, "fetchGlobalAchievementPercentages failed to decode JSON")
	}

	bytes, err := json.Marshal(result)
	if err == nil {
		if err := s.Cache.Set(ctx, cacheKey, bytes, time.Hour*24).Err(); err != nil {
			log.Printf("failed to cache GlobalAchievementPercentages: %v", err)
		}
	}

	return &result, nil
}
