package provider

import (
	"errors"
	"fmt"
	"net/url"
)

func sanitizeHTTPError(err error) string {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err.Error()
	}
	return err.Error()
}

type ImageGenerator interface {
	GenerateImage(prompt string, opts ImageOptions) (*ImageResult, error)
	Name() string
	ID() string
	Provider() string
	IsAvailable() bool
}

type MultiImageGenerator interface {
	ImageGenerator
	GenerateMultiImage(prompt string, inputImages []string, outputCount int, opts ImageOptions) (*MultiImageResult, error)
	SupportsMultiImage() bool
}

type ImageOptions struct {
	AspectRatio string
	ImageSize   string
	InputImages []string
	MaskImage   string
}

type ImageResult struct {
	Data     string
	MimeType string
}

type MultiImageResult struct {
	Images   []ImageResult
	MimeType string
}

type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
}

var models = make(map[string]ImageGenerator)

func Register(id string, generator ImageGenerator) {
	models[id] = generator
}

func Get(id string) (ImageGenerator, error) {
	if g, ok := models[id]; ok {
		if !g.IsAvailable() {
			return nil, fmt.Errorf("model %s is not configured or unavailable", id)
		}
		return g, nil
	}
	return nil, fmt.Errorf("unknown image generation model %s", id)
}

func GetDefault() (ImageGenerator, error) {
	priority := []string{"gpt-image-2", "gemini-3-pro-image-preview", "doubao-seedream-4-5"}
	for _, id := range priority {
		if g, ok := models[id]; ok && g.IsAvailable() {
			return g, nil
		}
	}
	return nil, fmt.Errorf("no available image generation model")
}

var modelOrder = []string{
	"gpt-image-2",
	"gemini-3.1-flash-image-preview",
	"gemini-3-pro-image-preview",
	"doubao-seedream-4-5",
}

func ListAvailable() []ModelInfo {
	var result []ModelInfo
	seen := make(map[string]bool)
	for _, id := range modelOrder {
		if g, ok := models[id]; ok {
			result = append(result, ModelInfo{
				ID:          id,
				Name:        g.Name(),
				Provider:    g.Provider(),
				Description: "",
				Available:   g.IsAvailable(),
			})
			seen[id] = true
		}
	}
	for id, g := range models {
		if seen[id] {
			continue
		}
		result = append(result, ModelInfo{
			ID:          id,
			Name:        g.Name(),
			Provider:    g.Provider(),
			Description: "",
			Available:   g.IsAvailable(),
		})
	}
	return result
}
