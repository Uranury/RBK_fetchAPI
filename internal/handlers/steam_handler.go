package handlers

import (
	"github.com/Uranury/RBK_fetchAPI/internal/apperrors"
	"github.com/Uranury/RBK_fetchAPI/internal/services"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	steamService *services.SteamService
}

func NewUserHandler(steamService *services.SteamService) *UserHandler {
	return &UserHandler{steamService: steamService}
}

func (h *UserHandler) RespondWithError(c *gin.Context, err error) {
	if apiErr, ok := err.(*apperrors.APIError); ok {
		c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
	} else {
		c.JSON(500, gin.H{"error": "internal server error"})
	}
}

// GetSteamID godoc
// @Summary      Retrieve steamID under vanityID if it exists
// @Tags         steamProfile
// @Produce      json
// @Param        vanity query string true "Vanity URL"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure 	 404 {object} apperrors.APIError
// @Failure      500 {object} apperrors.APIError
// @Router       /steam_id [get]
func (h *UserHandler) GetVanityProfile(c *gin.Context) {
	vanity := c.Query("vanity")
	if vanity == "" {
		c.JSON(400, gin.H{"error": "vanity is required"})
		return
	}

	steamID, err := h.steamService.ResolveVanityURL(c.Request.Context(), vanity)
	if err != nil {
		h.RespondWithError(c, err)
		return
	}

	c.JSON(200, gin.H{"steamID": steamID})
}

// GetOwnedGames godoc
// @Summary 	 returns user's owned games
// @Tags 	 	 gamesInfo
// @Produce 	 json
// @Param 		 steamID query string true "Steam ID"
// @Success 	 200 {object} models.OwnedGamesResponse
// @Failure 	 400 {object} map[string]string
// @Failure 	 500 {object} apperrors.APIError
// @Router 		 /games [get]
func (h *UserHandler) GetOwnedGames(c *gin.Context) {
	steamID := c.Query("steamID")
	if steamID == "" {
		c.JSON(400, gin.H{"error": "steam_id is required"})
		return
	}

	ownedGames, err := h.steamService.GetOwnedGames(c.Request.Context(), steamID)
	if err != nil {
		h.RespondWithError(c, err)
		return
	}

	c.JSON(200, ownedGames)
}

// GetUserSummary godoc
// @Summary 	 returns general info about the user
// @Tags 	 	 steamProfile
// @Produce 	 json
// @Param 		 steamID query string true "Steam ID"
// @Success 	 200 {object} models.Summary
// @Failure 	 400 {object} map[string]string
// @Failure 	 404 {object} apperrors.APIError
// @Failure 	 500 {object} apperrors.APIError
// @Router 	 	 /summary [get]
func (h *UserHandler) GetUserSummary(c *gin.Context) {
	steamID := c.Query("steamID")
	if steamID == "" {
		c.JSON(400, gin.H{"error": "steam_id is required"})
		return
	}

	summary, err := h.steamService.GetPlayerSummaries(c.Request.Context(), steamID)
	if err != nil {
		h.RespondWithError(c, err)
		return
	}

	c.JSON(200, summary)
}

// GetUserAchievements
// @Summary 	 returns all the achievements the user have for a game with all the details
// @Tags 		 gamesInfo
// @Produce 	 json
// @Param 		 steamID query string true "Steam ID of the user"
// @Param 		 appID query string true "App ID of the game"
// @Success 	 200 {object} models.PlayerAchievements
// @Failure 	 400 {object} map[string]string
// @Failure 	 409 {object} apperrors.APIError
// @Failure 	 500 {object} apperrors.APIError
// @Router 		 /achievements [get]
func (h *UserHandler) GetUserAchievements(c *gin.Context) {
	steamID, appID := c.Query("steamID"), c.Query("appID")
	if steamID == "" || appID == "" {
		c.JSON(400, gin.H{"error": "steam_id and app_id are required"})
	}

	achievements, err := h.steamService.GetPlayerAchievements(c.Request.Context(), steamID, appID)
	if err != nil {
		h.RespondWithError(c, err)
		return
	}

	c.JSON(200, achievements)
}
