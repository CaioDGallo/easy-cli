package utils

import "strings"

func SanitizeClientName(clientName string) string {
	sanitized := strings.TrimSpace(clientName)
	sanitized = strings.ToLower(sanitized)
	return strings.ReplaceAll(sanitized, " ", "-")
}
