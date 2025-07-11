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

	steamID, err := h.steamService.ResolveVanityURL(vanity)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to resolve vanity URL"})
		return
	}

	c.JSON(200, gin.H{"steamID": steamID})
}
