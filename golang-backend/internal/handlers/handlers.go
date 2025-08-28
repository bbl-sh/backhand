package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/john221wick/golang-backend/internal/auth"
	"github.com/john221wick/golang-backend/internal/middleware"
)

// ==== PocketBase Configuration and Admin Authentication ====

var (
	// base URL without trailing slash
	pbBaseURL       = "http://127.0.0.1:8090"
	pbAdminEmail    = "bhushanbharat6958@gmail.com"
	pbAdminPassword = "password123@A"

	adminToken    string
	adminTokenMu  sync.Mutex
	adminTokenExp time.Time
)

// pbAdminAuthResp defines the structure for the admin login response.
type pbAdminAuthResp struct {
	Token string `json:"token"`
	Admin struct {
		Id    string `json:"id"`
		Email string `json:"email"`
	} `json:"admin"`
}

// getAdminToken retrieves a cached or new admin token for PocketBase API calls.
func getAdminToken() (string, error) {
	adminTokenMu.Lock()
	defer adminTokenMu.Unlock()

	// Refresh token every 30 minutes (simple heuristic).
	if adminToken != "" && time.Now().Before(adminTokenExp) {
		fmt.Println("[DEBUG] Using cached admin token.")
		return adminToken, nil
	}

	fmt.Println("[DEBUG] Admin token is expired or not present. Requesting a new one...")
	body := map[string]string{
		"identity": pbAdminEmail,
		"password": pbAdminPassword,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal admin auth body: %w", err)
	}

	req, err := http.NewRequest("POST", pbBaseURL+"/api/collections/_superusers/auth-with-password", bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("build admin auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("admin auth http: %w", err)
	}
	defer resp.Body.Close()

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read admin auth response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("[DEBUG] Admin auth failed! Status: %d, Body: %s\n", resp.StatusCode, string(rb))
		return "", fmt.Errorf("admin auth failed: %d %s", resp.StatusCode, string(rb))
	}

	var ar pbAdminAuthResp
	if err := json.Unmarshal(rb, &ar); err != nil {
		return "", fmt.Errorf("decode admin auth: %w", err)
	}

	adminToken = ar.Token
	adminTokenExp = time.Now().Add(30 * time.Minute)
	fmt.Println("[DEBUG] Successfully obtained new admin token.")
	return adminToken, nil
}

// upsertChallengeStageRecord creates or updates a record in the "challengeStages" collection.
// It uses email + challengeName as the unique key.
func upsertChallengeStageRecord(email, challengeId, challengeName string) error {
	fmt.Printf("[DEBUG] Upserting challenge record for Email: %s, ChallengeName: %s, ChallengeId: %s\n", email, challengeName, challengeId)

	token, err := getAdminToken()
	if err != nil {
		return fmt.Errorf("get admin token: %w", err)
	}

	// ✅ CORRECT FIELD NAMES: ChallengeName (capital C) and email
	filter := fmt.Sprintf(`email="%s" && ChallengeName="%s"`,
		strings.ReplaceAll(email, `"`, `\"`),
		strings.ReplaceAll(challengeName, `"`, `\"`),
	)

	// ✅ IMPROVED DEBUGGING: Show raw filter before escaping
	fmt.Printf("[DEBUG] Raw filter: %s\n", filter)
	escapedFilter := url.QueryEscape(filter)
	apiURL := fmt.Sprintf("%s/api/collections/challengeStages/records?filter=%s&perPage=1", pbBaseURL, escapedFilter)
	fmt.Printf("[DEBUG] Search URL: %s\n", apiURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("build search request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		// ✅ SHOW FULL ERROR RESPONSE
		return fmt.Errorf("search failed: %d %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			Id string `json:"id"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("decode search response: %w, body: %s", err, string(body))
	}

	// ✅ CORRECT FIELD NAMES IN PAYLOAD (capital C)
	payload := map[string]any{
		"email":         email,
		"ChallengeId":   challengeId,   // 👈 Capital 'C'
		"ChallengeName": challengeName, // 👈 Capital 'C'
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	method, requestURL := "POST", fmt.Sprintf("%s/api/collections/challengeStages/records", pbBaseURL)
	if len(result.Items) > 0 {
		method = "PATCH"
		requestURL = fmt.Sprintf("%s/api/collections/challengeStages/records/%s", pbBaseURL, result.Items[0].Id)
		fmt.Printf("[DEBUG] Updating existing record ID: %s\n", result.Items[0].Id)
	} else {
		fmt.Printf("[DEBUG] Creating new record\n")
	}

	req, err = http.NewRequest(method, requestURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("build %s request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s request failed: %w", method, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s failed: %d %s", method, resp.StatusCode, string(respBody))
	}

	fmt.Printf("[DEBUG] Successfully %sd record. Response: %s\n", method, string(respBody))
	return nil
}

// ==== API Request and Response Types ====

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type TestResponse struct {
	Success        bool   `json:"success"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	ActualOutput   string `json:"actual_output"`
	ExpectedOutput string `json:"expected_output"`
	UserEmail      string `json:"user_email"`
	Timestamp      string `json:"timestamp"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// ==== HTTP Handlers ====

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqBody LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: "Invalid request body"})
		return
	}

	if reqBody.Email == "" || reqBody.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: "Email and password are required"})
		return
	}

	authData, err := auth.AuthenticateWithPocketBase(reqBody.Email, reqBody.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: "Authentication failed: " + err.Error()})
		return
	}

	response := LoginResponse{
		Success: true,
		Token:   authData.Token,
		Email:   authData.Email,
		Message: "Login successful",
	}

	json.NewEncoder(w).Encode(response)
}

type TestResult struct {
	Success        bool
	Status         string
	ActualOutput   string
	ExpectedOutput string
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authData := middleware.GetAuthData(r)
	if authData == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: "Authentication required"})
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: "Failed to parse form"})
		return
	}

	var problemID int
	_, err := fmt.Sscanf(r.FormValue("problemId"), "%d", &problemID)
	if err != nil || problemID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: "Invalid or missing problemId"})
		return
	}

	// Read challengeId and challengeName from form
	challengeId := r.FormValue("challengeId") // e.g., "1", "ch01", etc.
	challengeName := r.FormValue("challengeName")

	file, _, err := r.FormFile("code")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: "Missing code file"})
		return
	}
	defer file.Close()
	codeBytes, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: "Failed to read code file"})
		return
	}

	// Run the test
	result := RunTest(TestInput{
		ProblemID: problemID,
		Code:      codeBytes,
	})

	// On success, upsert the challenge stage record
	if result.Success && result.Status == "Accepted" {
		fmt.Printf("[DEBUG] Test accepted for user %s. Upserting challenge record.\n", authData.Email)
		if err := upsertChallengeStageRecord(authData.Email, challengeId, challengeName); err != nil {
			fmt.Printf("[pb] UPSERT FAILED: email=%s challengeName=%s err=%v\n", authData.Email, challengeName, err)
		} else {
			fmt.Printf("[pb] UPSERT OK: email=%s challengeName=%s\n", authData.Email, challengeName)
		}
	}

	message := "Execution completed"
	if !result.Success {
		message = result.Status
	}

	response := TestResponse{
		Success:        result.Success,
		Status:         result.Status,
		Message:        message,
		ActualOutput:   result.ActualOutput,
		ExpectedOutput: result.ExpectedOutput,
		UserEmail:      authData.Email,
		Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Simple ping handler (auth required)
func TestConnectionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authData := middleware.GetAuthData(r)
	if authData == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: "Authentication required"})
		return
	}

	response := map[string]any{
		"success":   true,
		"message":   "Connection verified successfully",
		"user":      authData.Email,
		"token":     "valid",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
