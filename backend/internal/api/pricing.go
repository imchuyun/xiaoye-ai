package api

import (
	"net/http"

	"google-ai-proxy/internal/db"
	"google-ai-proxy/internal/provider"

	"github.com/gin-gonic/gin"
)

const (
	ModelNanobanana  = "gemini-3-pro-image-preview"
	ModelNanobanana2 = "gemini-3.1-flash-image-preview"
	ModelSeedream45  = "doubao-seedream-4-5"
	ModelGPTImage2   = "gpt-image-2"
	ModelSeedance15  = "doubao-seedance-1-5-pro-251215"
	ModelVeo31       = "veo-3.1-generate-preview"
)

var ModelDisplayNames = map[string]string{
	ModelNanobanana:  "Nanobanana Pro",
	ModelNanobanana2: "Nanobanana 2",
	ModelSeedream45:  "Seedream-4.5",
	ModelGPTImage2:   "GPT Image 2",
	ModelSeedance15:  "Seedance-1.5",
	ModelVeo31:       "Veo 3.1",
}

func GetModelDisplayName(model string) string {
	if name, ok := ModelDisplayNames[model]; ok {
		return name
	}
	return model
}

var ImagePricingConfig = map[string]map[string]int{
	ModelNanobanana: {
		"1K": 10,
		"2K": 12,
		"4K": 20,
	},
	ModelNanobanana2: {
		"0.5K": 3,
		"1K":   6,
		"2K":   8,
		"4K":   12,
	},
	ModelSeedream45: {
		"2K": 6,
		"4K": 10,
	},
	ModelGPTImage2: {
		"1K": 10,
		"2K": 12,
		"4K": 20,
	},
}

var VideoPricingConfig = struct {
	BasePerSecond   map[string]int
	AudioMultiplier float64
}{
	BasePerSecond: map[string]int{
		"480p":  6,
		"720p":  10,
		"1080p": 16,
	},
	AudioMultiplier: 1.2,
}

const DefaultEcommerceModel = ModelSeedream45

func GetImageCredits(model, size string) int {
	modelPricing, ok := ImagePricingConfig[model]
	if !ok {
		return 10
	}

	credits, ok := modelPricing[size]
	if !ok {
		maxCredits := 0
		for _, v := range modelPricing {
			if v > maxCredits {
				maxCredits = v
			}
		}
		if maxCredits > 0 {
			return maxCredits
		}
		return 10
	}
	return credits
}

func GetEcommerceCredits(size string, count int) int {
	creditsPerImage := GetImageCredits(DefaultEcommerceModel, size)
	return creditsPerImage * count
}

func GetPricing(c *gin.Context) {
	imagePricing := []gin.H{
		imagePricingEntry(ModelNanobanana, "Gemini 3 Pro image model", []gin.H{
			{"size": "1K", "credits": ImagePricingConfig[ModelNanobanana]["1K"], "description": "1024x1024"},
			{"size": "2K", "credits": ImagePricingConfig[ModelNanobanana]["2K"], "description": "2048x2048"},
			{"size": "4K", "credits": ImagePricingConfig[ModelNanobanana]["4K"], "description": "4096x4096"},
		}),
		imagePricingEntry(ModelNanobanana2, "Gemini 3.1 Flash image model", []gin.H{
			{"size": "0.5K", "credits": ImagePricingConfig[ModelNanobanana2]["0.5K"], "description": "512x512"},
			{"size": "1K", "credits": ImagePricingConfig[ModelNanobanana2]["1K"], "description": "1024x1024"},
			{"size": "2K", "credits": ImagePricingConfig[ModelNanobanana2]["2K"], "description": "2048x2048"},
			{"size": "4K", "credits": ImagePricingConfig[ModelNanobanana2]["4K"], "description": "4096x4096"},
		}),
		imagePricingEntry(ModelSeedream45, "Volcengine Seedream image model", []gin.H{
			{"size": "2K", "credits": ImagePricingConfig[ModelSeedream45]["2K"], "description": "2048x2048"},
			{"size": "4K", "credits": ImagePricingConfig[ModelSeedream45]["4K"], "description": "4096x4096"},
		}),
		imagePricingEntry(ModelGPTImage2, "OpenAI image generation and editing", []gin.H{
			{"size": "1K", "credits": ImagePricingConfig[ModelGPTImage2]["1K"], "description": "1024x1024"},
			{"size": "2K", "credits": ImagePricingConfig[ModelGPTImage2]["2K"], "description": "1536x1024 / 1024x1536"},
			{"size": "4K", "credits": ImagePricingConfig[ModelGPTImage2]["4K"], "description": "high quality"},
		}),
	}

	enabledImagePricing := make([]gin.H, 0, len(imagePricing))
	for _, entry := range imagePricing {
		model, _ := entry["model"].(string)
		if db.IsModelEnabled(model) {
			enabledImagePricing = append(enabledImagePricing, entry)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"image": enabledImagePricing,
		"video": gin.H{
			"base_per_second":     VideoPricingConfig.BasePerSecond,
			"audio_multiplier":    VideoPricingConfig.AudioMultiplier,
			"veo_base_per_second": provider.VeoCreditsPerSecond,
		},
		"ecommerce": gin.H{
			"model":      DefaultEcommerceModel,
			"model_name": GetModelDisplayName(DefaultEcommerceModel),
			"prices":     ImagePricingConfig[DefaultEcommerceModel],
		},
		"exchange_rate": "1=10",
	})
}

func imagePricingEntry(model, description string, prices []gin.H) gin.H {
	return gin.H{
		"model":       model,
		"model_name":  GetModelDisplayName(model),
		"description": description,
		"prices":      prices,
	}
}
