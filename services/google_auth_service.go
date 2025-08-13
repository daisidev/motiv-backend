package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleAuthService struct {
	clientID     string
	clientSecret string
	config       *oauth2.Config
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func NewGoogleAuthService() *GoogleAuthService {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return &GoogleAuthService{
		clientID:     clientID,
		clientSecret: clientSecret,
		config:       config,
	}
}

// VerifyIDToken verifies the Google ID token and returns user information
func (g *GoogleAuthService) VerifyIDToken(idToken string) (*GoogleUserInfo, error) {
	// Create OAuth2 service
	ctx := context.Background()
	oauth2Service, err := oauth2.NewService(ctx, option.WithAPIKey(""))
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 service: %v", err)
	}

	// Verify the token
	tokenInfo, err := oauth2Service.Tokeninfo().IdToken(idToken).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %v", err)
	}

	// Check if the token is for our application
	if tokenInfo.Audience != g.clientID {
		return nil, fmt.Errorf("token audience mismatch")
	}

	// Check if token is expired
	if tokenInfo.ExpiresIn <= 0 {
		return nil, fmt.Errorf("token has expired")
	}

	// Extract user information
	userInfo := &GoogleUserInfo{
		ID:            tokenInfo.UserId,
		Email:         tokenInfo.Email,
		VerifiedEmail: tokenInfo.VerifiedEmail,
	}

	// Get additional user info if available
	if tokenInfo.Email != "" {
		// Try to get more detailed user info
		if detailedInfo, err := g.getUserInfo(idToken); err == nil {
			userInfo.Name = detailedInfo.Name
			userInfo.GivenName = detailedInfo.GivenName
			userInfo.FamilyName = detailedInfo.FamilyName
			userInfo.Picture = detailedInfo.Picture
		}
	}

	return userInfo, nil
}

// getUserInfo gets detailed user information using the access token
func (g *GoogleAuthService) getUserInfo(accessToken string) (*GoogleUserInfo, error) {
	ctx := context.Background()
	
	// Create a token source
	token := &oauth2.Token{AccessToken: accessToken}
	tokenSource := g.config.TokenSource(ctx, token)
	
	// Create OAuth2 service with the token
	oauth2Service, err := oauth2.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 service: %v", err)
	}

	// Get user info
	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	return &GoogleUserInfo{
		ID:            userInfo.Id,
		Email:         userInfo.Email,
		VerifiedEmail: userInfo.VerifiedEmail,
		Name:          userInfo.Name,
		GivenName:     userInfo.GivenName,
		FamilyName:    userInfo.FamilyName,
		Picture:       userInfo.Picture,
	}, nil
}

// ParseJWT parses a JWT token to extract claims (for client-side verification)
// Note: This is a simple parser and should not be used for security-critical verification
func (g *GoogleAuthService) ParseJWT(tokenString string) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode the payload (second part)
	payload := parts[1]
	
	// Add padding if necessary
	for len(payload)%4 != 0 {
		payload += "="
	}

	// Base64 decode
	decoded, err := base64URLDecode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %v", err)
	}

	// Parse JSON
	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %v", err)
	}

	return claims, nil
}

// base64URLDecode decodes base64url encoded string
func base64URLDecode(s string) ([]byte, error) {
	// Replace URL-safe characters
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	
	// Add padding
	for len(s)%4 != 0 {
		s += "="
	}
	
	return base64.StdEncoding.DecodeString(s)
}

// GenerateUsername generates a unique username from the user's name and email
func (g *GoogleAuthService) GenerateUsername(name, email string) string {
	// Start with the name, remove spaces and convert to lowercase
	username := strings.ToLower(strings.ReplaceAll(name, " ", ""))
	
	// If name is empty or too short, use email prefix
	if len(username) < 3 {
		emailParts := strings.Split(email, "@")
		if len(emailParts) > 0 {
			username = strings.ToLower(emailParts[0])
		}
	}
	
	// Remove any non-alphanumeric characters except underscores
	var cleanUsername strings.Builder
	for _, char := range username {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			cleanUsername.WriteRune(char)
		}
	}
	
	result := cleanUsername.String()
	
	// Ensure minimum length
	if len(result) < 3 {
		result = "user" + result
	}
	
	// Ensure maximum length
	if len(result) > 30 {
		result = result[:30]
	}
	
	return result
}