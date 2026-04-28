package provider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"google-ai-proxy/internal/config"
)

const openAIImageModelID = "gpt-image-2"

type OpenAIImageModel struct{}

func init() {
	Register(openAIImageModelID, &OpenAIImageModel{})
}

func (o *OpenAIImageModel) ID() string {
	return openAIImageModelID
}

func (o *OpenAIImageModel) Name() string {
	return "GPT Image 2"
}

func (o *OpenAIImageModel) Provider() string {
	return "openai"
}

func (o *OpenAIImageModel) IsAvailable() bool {
	return config.GetOpenAIAPIKey() != ""
}

func (o *OpenAIImageModel) GenerateImage(prompt string, opts ImageOptions) (*ImageResult, error) {
	if config.GetOpenAIAPIKey() == "" {
		return nil, fmt.Errorf("OpenAI API key is not configured")
	}
	if len(opts.InputImages) > 0 || opts.MaskImage != "" {
		return o.editImage(prompt, opts)
	}
	return o.generateImage(prompt, opts)
}

func (o *OpenAIImageModel) generateImage(prompt string, opts ImageOptions) (*ImageResult, error) {
	reqBody := openAIImageGenerationRequest{
		Model:   openAIImageModelID,
		Prompt:  prompt,
		Size:    openAIImageSize(opts.AspectRatio),
		Quality: openAIImageQuality(opts.ImageSize),
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", strings.TrimRight(config.GetOpenAIBaseURL(), "/")+"/images/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.GetOpenAIAPIKey())

	return o.doRequest(req)
}

func (o *OpenAIImageModel) editImage(prompt string, opts ImageOptions) (*ImageResult, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("model", openAIImageModelID)
	_ = writer.WriteField("prompt", prompt)
	_ = writer.WriteField("size", openAIImageSize(opts.AspectRatio))
	_ = writer.WriteField("quality", openAIImageQuality(opts.ImageSize))

	for i, img := range opts.InputImages {
		if err := writeBase64File(writer, "image", fmt.Sprintf("image-%d.png", i+1), img); err != nil {
			return nil, err
		}
	}
	if opts.MaskImage != "" {
		if err := writeBase64File(writer, "mask", "mask.png", opts.MaskImage); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", strings.TrimRight(config.GetOpenAIBaseURL(), "/")+"/images/edits", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+config.GetOpenAIAPIKey())

	return o.doRequest(req)
}

func (o *OpenAIImageModel) doRequest(req *http.Request) (*ImageResult, error) {
	client := &http.Client{Timeout: 600 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI image request failed: %s", sanitizeHTTPError(err))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("OpenAI image request failed: %s", parseOpenAIError(body, resp.StatusCode))
	}

	var parsed openAIImageResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("OpenAI image response parse failed: %v", err)
	}
	if len(parsed.Data) == 0 || parsed.Data[0].B64JSON == "" {
		return nil, fmt.Errorf("OpenAI image response did not include image data")
	}

	return &ImageResult{
		Data:     parsed.Data[0].B64JSON,
		MimeType: "image/png",
	}, nil
}

func writeBase64File(writer *multipart.Writer, fieldName, fileName, raw string) error {
	data := strings.TrimSpace(raw)
	if idx := strings.Index(data, ","); idx >= 0 && strings.HasPrefix(data[:idx+1], "data:") {
		data = data[idx+1:]
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("invalid base64 image: %w", err)
	}
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return err
	}
	_, err = part.Write(decoded)
	return err
}

func openAIImageSize(aspectRatio string) string {
	switch strings.TrimSpace(aspectRatio) {
	case "16:9", "4:3", "3:2", "5:4", "21:9":
		return "1536x1024"
	case "9:16", "3:4", "2:3", "4:5":
		return "1024x1536"
	default:
		return "1024x1024"
	}
}

func openAIImageQuality(imageSize string) string {
	switch strings.TrimSpace(imageSize) {
	case "4K":
		return "high"
	case "1K":
		return "medium"
	default:
		return "auto"
	}
}

func parseOpenAIError(body []byte, statusCode int) string {
	var parsed struct {
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error != nil && parsed.Error.Message != "" {
		return parsed.Error.Message
	}
	return fmt.Sprintf("status %d: %s", statusCode, string(body))
}

type openAIImageGenerationRequest struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Size    string `json:"size,omitempty"`
	Quality string `json:"quality,omitempty"`
}

type openAIImageResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
		URL     string `json:"url"`
	} `json:"data"`
}
