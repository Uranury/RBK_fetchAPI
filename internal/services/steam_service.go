package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Uranury/RBK_fetchAPI/internal/models"
	"github.com/Uranury/RBK_fetchAPI/internal/repositories"
	"github.com/redis/go-redis/v9"
)

// TODO: inject and reuse an http.Client with timeout
// TODO: replace all http.Get calls with http.NewRequestWithContext + client.Do
// TODO: add request history saving consistently across all SteamService methods
// TODO: validate Steam API response bodies for "success": false and other edge cases
// TODO: improve error wrapping/logging for JSON decoding and API failures
// TODO: extract hardcoded URLs into constants
type SteamService struct {
	APIKey    string
	Cache     *redis.Client
	steamRepo repositories.SteamRepository
}

func NewSteamService(APIKey string, Cache *redis.Client, steamRepo repositories.SteamRepository) *SteamService {
	return &SteamService{APIKey: APIKey, Cache: Cache, steamRepo: steamRepo}
}

func (s *SteamService) ResolveVanityURL(ctx context.Context, vanityName string) (string, error) {
	start := time.Now()
	cacheKey := fmt.Sprintf("vanity:%s", vanityName)
	params := map[string]interface{}{"vanityName": vanityName}

	if steamID, err := s.Cache.Get(ctx, cacheKey).Result(); err == nil {
		_ = s.steamRepo.SaveRequestHistory("/steam_id:ResolveVanityURL", params, true, "", time.Since(start))
		return steamID, nil
	}

	url := fmt.Sprintf("http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?key=%s&vanityurl=%s", s.APIKey, vanityName)

	resp, err := http.Get(url)
	if err != nil {
		_ = s.steamRepo.SaveRequestHistory("/steam_id:ResolveVanityURL", params, false, err.Error(), time.Since(start))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("SteamAPI responded with status: %s", resp.Status)
		_ = s.steamRepo.SaveRequestHistory("/steam_id:ResolveVanityURL", params, false, msg, time.Since(start))
		return "", errors.New(msg)
	}

	var result struct {
		Response struct {
			SteamID string `json:"steamid"`
			Success int    `json:"success"`
			Message string `json:"message,omitempty"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		_ = s.steamRepo.SaveRequestHistory("/steam_id:ResolveVanityURL", params, false, err.Error(), time.Since(start))
		return "", err
	}

	if result.Response.Success != 1 {
		msg := fmt.Sprintf("could not resolve vanity URL: %s", result.Response.Message)
		_ = s.steamRepo.SaveRequestHistory("/steam_id:ResolveVanityURL", params, false, msg, time.Since(start))
		return "", fmt.Errorf("could not resolve vanity URL: %s", result.Response.Message)
	}

	if err := s.Cache.Set(ctx, cacheKey, result.Response.SteamID, time.Minute*5).Err(); err != nil {
		log.Println("failed to cache vanity")
	}
	_ = s.steamRepo.SaveRequestHistory("/steam_id:ResolveVanityURL", params, true, "", time.Since(start))
	return result.Response.SteamID, nil
}

func (s *SteamService) GetOwnedGames(ctx context.Context, steamID string) (*models.OwnedGamesResponse, error) {
	cacheKey := fmt.Sprintf("owned_games:%s", steamID)
	cached, err := s.Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var games models.OwnedGamesResponse
		if err := json.Unmarshal([]byte(cached), &games); err == nil {
			return &games, nil
		}
	}

	url := fmt.Sprintf(
		"https://api.steampowered.com/IPlayerService/GetOwnedGames/v1/?key=%s&steamid=%s&include_appinfo=true",
		s.APIKey,
		steamID,
	)
	var response models.OwnedGamesResponse

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam API responded with status: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	bytes, err := json.Marshal(response)
	if err == nil {
		s.Cache.Set(ctx, cacheKey, bytes, time.Minute*5)
	}

	return &response, nil
}

func (s *SteamService) GetPlayerSummaries(ctx context.Context, steamID string) (*models.Summary, error) {
	cacheKey := fmt.Sprintf("summary:%s", steamID)

	cached, err := s.Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var summary models.Summary
		if err := json.Unmarshal([]byte(cached), &summary); err == nil {
			return &summary, nil
		}
	}

	url := fmt.Sprintf(
		"http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		s.APIKey,
		steamID,
	)
	var result models.Summary

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam API responded with status: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	bytes, err := json.Marshal(result)
	if err == nil {
		s.Cache.Set(ctx, cacheKey, bytes, time.Minute*5)
	}

	return &result, nil
}

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

	// Create a map for quick lookup of achievement details
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

	// Combine player achievements with schema data
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

	url := fmt.Sprintf("http://api.steampowered.com/ISteamUserStats/GetPlayerAchievements/v0001/?appid=%s&key=%s&steamid=%s", appID, s.APIKey, steamID)

	resp, err := http.Get(url)
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

	url := fmt.Sprintf("http://api.steampowered.com/ISteamUserStats/GetSchemaForGame/v2/?key=%s&appid=%s", s.APIKey, appID)

	resp, err := http.Get(url)
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
