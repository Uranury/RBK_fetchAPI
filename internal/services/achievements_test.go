package services_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/Uranury/RBK_fetchAPI/internal/models"
	"github.com/Uranury/RBK_fetchAPI/internal/services"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// RedisClientInterface defines the interface we need from Redis client
type RedisClientInterface interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

// MockRedisClient is a mock implementation of Redis client interface
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

// MockSteamRepository is a mock implementation of SteamRepository
type MockSteamRepository struct {
	mock.Mock
}

func (m *MockSteamRepository) SaveRequestHistory(endpoint string, params map[string]interface{}, success bool, errorMsg string, duration time.Duration) error {
	args := m.Called(endpoint, params, success, errorMsg, duration)
	return args.Error(0)
}

// TestSteamService wraps the actual service to allow dependency injection
type TestSteamService struct {
	*services.SteamService
	redisClient RedisClientInterface
	baseURL     string // Add this to override the API URLs
}

func NewTestSteamService(apiKey string, redisClient RedisClientInterface, repo *MockSteamRepository, httpClient *http.Client, baseURL string) *TestSteamService {
	// Create a real service with a real Redis client (we'll override the methods we need)
	realRedisClient := redis.NewClient(&redis.Options{})
	service := services.NewSteamService(apiKey, realRedisClient, repo, httpClient)

	return &TestSteamService{
		SteamService: service,
		redisClient:  redisClient,
		baseURL:      baseURL,
	}
}

type SteamServiceTestSuite struct {
	suite.Suite
	service     *TestSteamService
	redisMock   *MockRedisClient
	repoMock    *MockSteamRepository
	server      *httptest.Server
	testContext context.Context
}

// TestSteamService wraps the actual service to allow dependency injection

// Override the cache methods to use our mock
func (ts *TestSteamService) GetFromCache(ctx context.Context, key string) *redis.StringCmd {
	return ts.redisClient.Get(ctx, key)
}

func (ts *TestSteamService) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return ts.redisClient.Set(ctx, key, value, expiration)
}

// In your SetupTest method, update this part:
func (suite *SteamServiceTestSuite) SetupTest() {
	suite.redisMock = new(MockRedisClient)
	suite.repoMock = new(MockSteamRepository)
	suite.testContext = context.Background()

	// Setup test HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ... your existing handler code ...
	}))

	// Create HTTP client that uses the test server
	httpClient := &http.Client{
		Transport: &http.Transport{
			// You might need to create a custom transport that redirects calls to your test server
		},
	}

	// Create service with mocked dependencies
	suite.service = NewTestSteamService(
		"test_api_key",
		suite.redisMock,
		suite.repoMock,
		httpClient,
		suite.server.URL, // Pass the test server URL
	)
}

// Fix your test expectations:
func (suite *SteamServiceTestSuite) TestCacheHit() {
	expected := &models.PlayerAchievements{
		SteamID:  "76561197960434622",
		GameName: "Test Game",
	}
	cachedData, _ := json.Marshal(expected)

	// Mock cache hit
	suite.redisMock.On("Get", suite.testContext, "player_achievements:76561197960434622:game:123").
		Return(redis.NewStringResult(string(cachedData), nil))

	// Expect logging call - THIS IS THE KEY FIX
	suite.repoMock.On("SaveRequestHistory",
		"/achievements:GetPlayerAchievements",
		map[string]interface{}{"steamID": "76561197960434622", "appID": "123"},
		true, // success should be true for cache hit
		"",   // empty error message for success
		mock.AnythingOfType("time.Duration"),
	).Return(nil)

	result, apiErr := suite.service.GetPlayerAchievements(suite.testContext, "76561197960434622", "123")

	suite.NoError(apiErr)
	suite.Equal(expected, result)
	suite.redisMock.AssertExpectations(suite.T())
	suite.repoMock.AssertExpectations(suite.T())
}

// For TestCacheMissSuccess, you need to ensure your HTTP calls succeed
func (suite *SteamServiceTestSuite) TestCacheMissSuccess() {
	// Mock cache miss for main key
	suite.redisMock.On("Get", suite.testContext, "player_achievements:76561197960434622:game:123").
		Return(redis.NewStringResult("", redis.Nil))

	// Mock cache misses for helper endpoints
	suite.redisMock.On("Get", suite.testContext, "fetched_player_achievements:76561197960434622:game:123").
		Return(redis.NewStringResult("", redis.Nil))
	suite.redisMock.On("Get", suite.testContext, "game_schema:123").
		Return(redis.NewStringResult("", redis.Nil))
	suite.redisMock.On("Get", suite.testContext, "global_achievement_percentages:123").
		Return(redis.NewStringResult("", redis.Nil))

	// Mock cache sets
	suite.redisMock.On("Set", suite.testContext, mock.Anything, mock.Anything, mock.Anything).
		Return(redis.NewStatusResult("OK", nil)).Times(4)

	// THE KEY FIX: Expect the actual success case logging
	suite.repoMock.On("SaveRequestHistory",
		"/achievements:GetPlayerAchievements",
		map[string]interface{}{"steamID": "76561197960434622", "appID": "123"},
		true, // Should be true for successful API call
		"",   // Empty error message for success
		mock.AnythingOfType("time.Duration"),
	).Return(nil)

	result, apiErr := suite.service.GetPlayerAchievements(suite.testContext, "76561197960434622", "123")

	suite.NoError(apiErr)
	suite.Equal("76561197960434622", result.SteamID)
	suite.Equal("Test Game", result.GameName)
	suite.Len(result.Achievements, 2)

	// Verify achievement details
	winAch := result.Achievements[0]
	suite.Equal("ACH_WIN", winAch.Name)
	suite.True(winAch.Achieved)
	suite.Equal(25.5, winAch.Rarity)
	suite.Equal(time.Unix(1666666666, 0), winAch.UnlockTime)

	loseAch := result.Achievements[1]
	suite.Equal("ACH_LOSE", loseAch.Name)
	suite.False(loseAch.Achieved)
	suite.Equal(75.0, loseAch.Rarity)
	suite.Zero(loseAch.UnlockTime)
}
