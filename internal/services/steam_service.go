package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Uranury/RBK_fetchAPI/internal/repositories"
	"github.com/redis/go-redis/v9"
)

type SteamService struct {
	APIKey    string
	Cache     *redis.Client
	steamRepo repositories.SteamRepository
}

func NewSteamService(APIKey string, Cache *redis.Client, steamRepo repositories.SteamRepository) *SteamService {
	return &SteamService{APIKey: APIKey, Cache: Cache, steamRepo: steamRepo}
}

func (s *SteamService) ResolveVanityURL(vanityName string) (string, error) {
	start := time.Now()
	ctx := context.Background()
	cacheKey := "vanity:" + vanityName
	params := map[string]interface{}{"vanityName": vanityName}

	if steamID, err := s.Cache.Get(ctx, cacheKey).Result(); err == nil {
		_ = s.steamRepo.SaveRequestHistory("ResolveVanityURL", params, true, "", time.Since(start))
		return steamID, nil
	}

	url := fmt.Sprintf("http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?key=%s&vanityurl=%s", s.APIKey, vanityName)

	resp, err := http.Get(url)
	if err != nil {
		_ = s.steamRepo.SaveRequestHistory("ResolveVanityURL", params, false, err.Error(), time.Since(start))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("SteamAPI responded with status: %s", resp.Status)
		_ = s.steamRepo.SaveRequestHistory("ResolveVanityURL", params, false, msg, time.Since(start))
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
		_ = s.steamRepo.SaveRequestHistory("ResolveVanityURL", params, false, err.Error(), time.Since(start))
		return "", err
	}

	if result.Response.Success != 1 {
		msg := fmt.Sprintf("could not resolve vanity URL: %s", result.Response.Message)
		_ = s.steamRepo.SaveRequestHistory("ResolveVanityURL", params, false, msg, time.Since(start))
		return "", fmt.Errorf("could not resolve vanity URL: %s", result.Response.Message)
	}

	if err := s.Cache.Set(ctx, cacheKey, result.Response.SteamID, time.Minute*5).Err(); err != nil {
		log.Println("failed to cache vanity")
	}
	_ = s.steamRepo.SaveRequestHistory("ResolveVanityURL", params, true, "", time.Since(start))
	return result.Response.SteamID, nil
}
