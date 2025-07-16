package handlers

import (
	"github.com/Uranury/RBK_fetchAPI/internal/services"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	steamService *services.SteamService
}

func NewUserHandler(steamService *services.SteamService) *UserHandler {
	return &UserHandler{steamService: steamService}
}

// TODO: edit swagger so that in success it returns not a map, but the model

// GetSteamID godoc
// @Summary      Retrieve steamID under vanityID if it exists
// @Tags         steamProfile
// @Produce      json
// @Param        vanity query string true "Vanity URL"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /steam_id [get]
func (h *UserHandler) GetVanityProfile(c *gin.Context) {
	vanity := c.Query("vanity")
	if vanity == "" {
		c.JSON(400, gin.H{"error": "vanity is required"})
		return
	}

	steamID, err := h.steamService.ResolveVanityURL(c.Request.Context(), vanity)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"steamID": steamID})
}

func (h *UserHandler) GetOwnedGames(c *gin.Context) {
	steamID := c.Query("steamID")
	if steamID == "" {
		c.JSON(400, gin.H{"error": "steam_id is required"})
		return
	}

	ownedGames, err := h.steamService.GetOwnedGames(c.Request.Context(), steamID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, ownedGames)
}

func (h *UserHandler) GetUserSummary(c *gin.Context) {
	steamID := c.Query("steamID")
	if steamID == "" {
		c.JSON(400, gin.H{"error": "steam_id is required"})
		return
	}

	summary, err := h.steamService.GetPlayerSummaries(c.Request.Context(), steamID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, summary)
}

func (h *UserHandler) GetUserAchievements(c *gin.Context) {
	steamID, appID := c.Query("steamID"), c.Query("appID")
	if steamID == "" || appID == "" {
		c.JSON(400, gin.H{"error": "steam_id and app_id are required"})
	}

	achievements, err := h.steamService.GetPlayerAchievements(c.Request.Context(), steamID, appID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, achievements)
}

// checking
