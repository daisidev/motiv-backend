package utils

import (
	"regexp"
	"strings"
)

// GenerateUsername creates a username from a name and email
func GenerateUsername(name, email string) string {
	// Try to use the name first
	if name != "" {
		username := normalizeUsername(name)
		if username != "" {
			return username
		}
	}

	// Fall back to email prefix
	emailParts := strings.Split(email, "@")
	if len(emailParts) > 0 {
		return normalizeUsername(emailParts[0])
	}

	return "user"
}

// normalizeUsername cleans up a string to be a valid username
func normalizeUsername(input string) string {
	// Convert to lowercase
	username := strings.ToLower(input)

	// Remove special characters, keep only alphanumeric and underscore
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	username = reg.ReplaceAllString(username, "")

	// Remove leading/trailing underscores
	username = strings.Trim(username, "_")

	// Ensure minimum length
	if len(username) < 3 {
		return ""
	}

	// Ensure maximum length
	if len(username) > 20 {
		username = username[:20]
	}

	return username
}
