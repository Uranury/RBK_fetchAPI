package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Uranury/RBK_fetchAPI/internal/apperrors"
	"github.com/Uranury/RBK_fetchAPI/internal/models"
	"github.com/Uranury/RBK_fetchAPI/internal/repositories"
	"github.com/redis/go-redis/v9"
)

const (
	resolveVanityURLTemplate = "http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?key=%s&vanityurl=%s"
	ownedGamesURLTemplate    = "https://api.steampowered.com/IPlayerService/GetOwnedGames/v1/?key=%s&steamid=%s&include_appinfo=true"
	playerSummariesTemplate  = "http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s"
)

type SteamService struct {
	APIKey     string
	Cache      *redis.Client
	steamRepo  repositories.SteamRepository
	HTTPClient *http.Client
}

func NewSteamService(APIKey string, Cache *redis.Client, steamRepo repositories.SteamRepository, client *http.Client) *SteamService {
	return &SteamService{
		APIKey:     APIKey,
		Cache:      Cache,
		steamRepo:  steamRepo,
		HTTPClient: client,
	}
}

func (s *SteamService) logRequest(endpoint string, params map[string]interface{}, success bool, errorMsg string, duration time.Duration) {
	if err := s.steamRepo.SaveRequestHistory(endpoint, params, success, errorMsg, duration); err != nil {
		log.Printf("failed to save request history: %v", err)
	}
}

func (s *SteamService) ResolveVanityURL(ctx context.Context, vanityName string) (string, error) {
	start := time.Now()
	endpoint := "/steam_id:ResolveVanityURL"
	params := map[string]interface{}{"vanityName": vanityName}

	cacheKey := fmt.Sprintf("vanity:%s", vanityName)
	if steamID, err := s.Cache.Get(ctx, cacheKey).Result(); err == nil {
		s.logRequest(endpoint, params, true, "", time.Since(start))
		return steamID, nil
	}

	url := fmt.Sprintf(resolveVanityURLTemplate, s.APIKey, vanityName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return "", apperrors.WrapAPIError(500, err, "ResolveVanityURL request creation failed")
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return "", apperrors.WrapAPIError(500, err, "ResolveVanityURL request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("steam API responded with status: %s", resp.Status)
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return "", apperrors.WrapAPIError(resp.StatusCode, err, "Unexpected error")
	}

	var result struct {
		Response struct {
			SteamID string `json:"steamid"`
			Success int    `json:"success"`
			Message string `json:"message,omitempty"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return "", apperrors.WrapAPIError(500, err, "ResolveVanityURL JSON decode failed")
	}

	if result.Response.Success != 1 {
		err := fmt.Errorf("could not resolve vanity URL: %s", result.Response.Message)
		s.logRequest(endpoint, params, true, "", time.Since(start))
		return "", apperrors.WrapAPIError(404, err, "No match")
	}

	if err := s.Cache.Set(ctx, cacheKey, result.Response.SteamID, 5*time.Minute).Err(); err != nil {
		log.Printf("failed to cache vanity: %v", err)
	}

	s.logRequest(endpoint, params, true, "", time.Since(start))
	return result.Response.SteamID, nil
}

func (s *SteamService) GetOwnedGames(ctx context.Context, steamID string) (*models.OwnedGamesResponse, error) {
	start := time.Now()
	endpoint := "/games:GetOwnedGames"
	params := map[string]interface{}{"steam_id": steamID}

	cacheKey := fmt.Sprintf("owned_games:%s", steamID)
	cached, err := s.Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var games models.OwnedGamesResponse
		if err := json.Unmarshal([]byte(cached), &games); err == nil {
			s.logRequest(endpoint, params, true, "", time.Since(start))
			return &games, nil
		}
		log.Printf("failed to unmarshal cached games: %v", err)
	}

	url := fmt.Sprintf(ownedGamesURLTemplate, s.APIKey, steamID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return nil, apperrors.WrapAPIError(500, err, "GetOwnedGames request creation failed")
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return nil, apperrors.WrapAPIError(500, err, "GetOwnedGames API call failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("steam API responded with status: %s", resp.Status)
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return nil, apperrors.WrapAPIError(resp.StatusCode, err, "Steam GetOwnedGames returned non-200")
	}

	var response models.OwnedGamesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return nil, apperrors.WrapAPIError(500, err, "GetOwnedGames JSON decode failed")
	}

	if response.Response.GameCount == 0 {
		log.Printf("no games found for steamID: %s", steamID)
	}

	bytes, err := json.Marshal(response)
	if err == nil {
		if err := s.Cache.Set(ctx, cacheKey, bytes, 5*time.Minute).Err(); err != nil {
			log.Printf("failed to cache owned games: %v", err)
		}
	}

	s.logRequest(endpoint, params, true, "", time.Since(start))
	return &response, nil
}

func (s *SteamService) GetPlayerSummaries(ctx context.Context, steamID string) (*models.Summary, error) {
	start := time.Now()
	endpoint := "/summary:GetPlayerSummaries"
	params := map[string]interface{}{"steam_id": steamID}

	cacheKey := fmt.Sprintf("summary:%s", steamID)
	cached, err := s.Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var summary models.Summary
		if err := json.Unmarshal([]byte(cached), &summary); err == nil {
			s.logRequest(endpoint, params, true, "", time.Since(start))
			return &summary, nil
		}
		log.Printf("failed to unmarshal cached summary: %v", err)
	}

	url := fmt.Sprintf(playerSummariesTemplate, s.APIKey, steamID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return nil, apperrors.WrapAPIError(500, err, "GetPlayerSummaries request creation failed")
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return nil, apperrors.WrapAPIError(500, err, "GetPlayerSummaries API call failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("steam API responded with status: %s", resp.Status)
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return nil, apperrors.WrapAPIError(resp.StatusCode, err, "Steam GetPlayerSummaries returned non-200")
	}

	var result models.Summary
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logRequest(endpoint, params, false, err.Error(), time.Since(start))
		return nil, apperrors.WrapAPIError(500, err, "GetPlayerSummaries JSON decode failed")
	}

	if len(result.Response.Players) == 0 {
		err := fmt.Errorf("no player found for steamID: %s", steamID)
		s.logRequest(endpoint, params, true, err.Error(), time.Since(start))
		return nil, apperrors.WrapAPIError(404, err, "No player found")
	}

	bytes, err := json.Marshal(result)
	if err == nil {
		if err := s.Cache.Set(ctx, cacheKey, bytes, 5*time.Minute).Err(); err != nil {
			log.Printf("failed to cache player summary: %v", err)
		}
	}

	s.logRequest(endpoint, params, true, "", time.Since(start))
	return &result, nil
}
