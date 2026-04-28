package db

import (
	"time"

	"gorm.io/gorm/clause"
)

// ModelConfig stores admin-controlled model visibility.
type ModelConfig struct {
	ModelID   string    `gorm:"primaryKey;type:varchar(100);comment:Provider model ID" json:"model_id"`
	Enabled   bool      `gorm:"type:boolean;default:true;index;comment:Whether model is visible/selectable" json:"enabled"`
	CreatedAt time.Time `gorm:"type:datetime;comment:Creation time" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:datetime;comment:Last update time" json:"updated_at"`
}

// EnsureModelConfigs creates default enabled rows for registered models.
func EnsureModelConfigs(modelIDs []string) {
	if DB == nil || len(modelIDs) == 0 {
		return
	}

	now := time.Now()
	rows := make([]ModelConfig, 0, len(modelIDs))
	for _, id := range modelIDs {
		if id == "" {
			continue
		}
		rows = append(rows, ModelConfig{
			ModelID:   id,
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	if len(rows) == 0 {
		return
	}

	DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows)
}

func GetModelConfigMap(modelIDs []string) map[string]ModelConfig {
	result := make(map[string]ModelConfig, len(modelIDs))
	if DB == nil || len(modelIDs) == 0 {
		for _, id := range modelIDs {
			result[id] = ModelConfig{ModelID: id, Enabled: true}
		}
		return result
	}

	EnsureModelConfigs(modelIDs)

	var rows []ModelConfig
	DB.Where("model_id IN ?", modelIDs).Find(&rows)
	for _, row := range rows {
		result[row.ModelID] = row
	}
	for _, id := range modelIDs {
		if _, ok := result[id]; !ok {
			result[id] = ModelConfig{ModelID: id, Enabled: true}
		}
	}
	return result
}

func IsModelEnabled(modelID string) bool {
	if modelID == "" || DB == nil {
		return true
	}

	EnsureModelConfigs([]string{modelID})

	var row ModelConfig
	if err := DB.First(&row, "model_id = ?", modelID).Error; err != nil {
		return true
	}
	return row.Enabled
}
