package api

import (
	"strconv"
	"strings"
)

// GenerateEcommercePromptSuffix builds an optional suffix for ecommerce image prompts.
func GenerateEcommercePromptSuffix(imageType, ecommerceType string, outputCount int) string {
	var parts []string
	if outputCount > 1 {
		parts = append(parts, strconv.Itoa(outputCount)+" images")
	}
	if ecommerceType != "" {
		parts = append(parts, ecommerceType+" platform")
	}
	if imageType != "" {
		parts = append(parts, imageType)
	}
	if len(parts) == 0 {
		return ""
	}
	return ". Generate " + strings.Join(parts, " ")
}
