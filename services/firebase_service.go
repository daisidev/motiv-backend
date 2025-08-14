package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type FirebaseService struct {
	client *auth.Client
}

type FirebaseUserInfo struct {
	UID           string `json:"uid"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}

func NewFirebaseService() (*FirebaseService, error) {
	ctx := context.Background()

	// Get the service account key from environment variable
	serviceAccountKey := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY")
	if serviceAccountKey == "" {
		return nil, fmt.Errorf("FIREBASE_SERVICE_ACCOUNT_KEY environment variable not set")
	}

	// Parse the service account key
	var serviceAccount map[string]interface{}
	if err := json.Unmarshal([]byte(serviceAccountKey), &serviceAccount); err != nil {
		return nil, fmt.Errorf("failed to parse service account key: %v", err)
	}

	// Initialize Firebase app
	opt := option.WithCredentialsJSON([]byte(serviceAccountKey))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %v", err)
	}

	// Get Auth client
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase Auth client: %v", err)
	}

	return &FirebaseService{
		client: client,
	}, nil
}

func (fs *FirebaseService) VerifyIDToken(idToken string) (*FirebaseUserInfo, error) {
	ctx := context.Background()

	token, err := fs.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %v", err)
	}

	// Get user info
	userRecord, err := fs.client.GetUser(ctx, token.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user record: %v", err)
	}

	return &FirebaseUserInfo{
		UID:           userRecord.UID,
		Email:         userRecord.Email,
		Name:          userRecord.DisplayName,
		Picture:       userRecord.PhotoURL,
		EmailVerified: userRecord.EmailVerified,
	}, nil
}
