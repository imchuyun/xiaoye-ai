package api

import (
	"net/http"

	"google-ai-proxy/internal/db"
	"google-ai-proxy/internal/provider"

	"github.com/gin-gonic/gin"
)

// GetModels returns enabled and currently available image generation models.
func GetModels(c *gin.Context) {
	models := provider.ListAvailable()
	modelIDs := make([]string, 0, len(models))
	for _, m := range models {
		modelIDs = append(modelIDs, m.ID)
	}
	configs := db.GetModelConfigMap(modelIDs)

	type ModelResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Provider    string `json:"provider"`
		Description string `json:"description"`
		Available   bool   `json:"available"`
	}

	result := make([]ModelResponse, 0, len(models))
	for _, m := range models {
		if !configs[m.ID].Enabled || !m.Available {
			continue
		}

		displayName := GetModelDisplayName(m.ID)
		if displayName == m.ID {
			displayName = m.Name
		}

		result = append(result, ModelResponse{
			ID:          m.ID,
			Name:        displayName,
			Provider:    m.Provider,
			Description: m.Description,
			Available:   m.Available,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"models": result,
	})
}
