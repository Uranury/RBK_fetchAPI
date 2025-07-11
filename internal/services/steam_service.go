package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type SteamService struct {
	APIKey string
	Cache  *redis.Client
}

func NewSteamService(APIKey string, Cache *redis.Client) *SteamService {
	return &SteamService{APIKey: APIKey, Cache: Cache}
}

func (s *SteamService) ResolveVanityURL(vanityName string) (string, error) {
	ctx := context.Background()
	cacheKey := "vanity:" + vanityName

	if steamID, err := s.Cache.Get(ctx, cacheKey).Result(); err == nil {
		return steamID, nil
	}

	url := fmt.Sprintf("http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?key=%s&vanityurl=%s", s.APIKey, vanityName)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("SteamAPI responded with status: %s", resp.Status)
	}

	var result struct {
		Response struct {
			SteamID string `json:"steamid"`
			Success int    `json:"success"`
			Message string `json:"message,omitempty"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Response.Success != 1 {
		return "", fmt.Errorf("could not resolve vanity URL: %s", result.Response.Message)
	}

	if err := s.Cache.Set(ctx, cacheKey, result.Response.SteamID, time.Minute*5).Err(); err != nil {
		log.Println("failed to cache vanity")
	}
	return result.Response.SteamID, nil
}
