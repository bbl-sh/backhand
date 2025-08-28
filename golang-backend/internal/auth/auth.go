package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const PocketBaseURL = "http://127.0.0.1:8090"

type LoginRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type PocketBaseUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type LoginResponse struct {
	Token  string         `json:"token"`
	Record PocketBaseUser `json:"record"`
}

type AuthData struct {
	Token  string
	Email  string
	UserID string
}

// AuthenticateWithPocketBase authenticates user with PocketBase
func AuthenticateWithPocketBase(email, password string) (*AuthData, error) {
	loginData := LoginRequest{
		Identity: email,
		Password: password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login data: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Use the correct PocketBase auth endpoint
	resp, err := client.Post(
		fmt.Sprintf("%s/api/collections/users/auth-with-password", PocketBaseURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PocketBase: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed: status %d", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &AuthData{
		Token:  loginResp.Token,
		Email:  loginResp.Record.Email,
		UserID: loginResp.Record.ID,
	}, nil
}

// ValidateToken validates PocketBase token
func ValidateToken(token string) (*AuthData, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/collections/users/auth-refresh", PocketBaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid token: status %d", resp.StatusCode)
	}

	var authResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &AuthData{
		Token:  authResp.Token,
		Email:  authResp.Record.Email,
		UserID: authResp.Record.ID,
	}, nil
}
