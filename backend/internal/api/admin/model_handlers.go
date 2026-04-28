package admin

import (
	"net/http"
	"strings"
	"time"

	"google-ai-proxy/internal/api"
	"google-ai-proxy/internal/db"
	"google-ai-proxy/internal/provider"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

func ListModels(c *gin.Context) {
	models := provider.ListAvailable()
	modelIDs := make([]string, 0, len(models))
	for _, model := range models {
		modelIDs = append(modelIDs, model.ID)
	}
	configs := db.GetModelConfigMap(modelIDs)

	items := make([]adminModelResponse, 0, len(models))
	for _, model := range models {
		name := api.GetModelDisplayName(model.ID)
		if name == model.ID {
			name = model.Name
		}
		cfg := configs[model.ID]
		items = append(items, adminModelResponse{
			ID:        model.ID,
			Name:      name,
			Provider:  model.Provider,
			Available: model.Available,
			Enabled:   cfg.Enabled,
			UpdatedAt: cfg.UpdatedAt.UnixMilli(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func UpdateModel(c *gin.Context) {
	modelID := strings.TrimSpace(c.Param("model_id"))
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing model id"})
		return
	}

	known := false
	for _, model := range provider.ListAvailable() {
		if model.ID == modelID {
			known = true
			break
		}
	}
	if !known {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	var req updateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	now := time.Now()
	err := db.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "model_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"enabled":    req.Enabled,
			"updated_at": now,
		}),
	}).Create(&db.ModelConfig{
		ModelID:   modelID,
		Enabled:   req.Enabled,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update model"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"item": gin.H{
			"id":      modelID,
			"enabled": req.Enabled,
		},
	})
}
