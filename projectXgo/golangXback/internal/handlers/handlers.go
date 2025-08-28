package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"your-app/internal/judge"
	"your-app/internal/middleware"

	"github.com/pocketbase/pocketbase/sdk"
)

// App holds shared dependencies for handlers
type App struct {
	PB *sdk.Client
}

// NewApp returns App with given PocketBase SDK client
func NewApp(client *sdk.Client) *App {
	return &App{PB: client}
}

// LoginHandler handles POST /login
// body: { "email":"...", "password":"..." }
func (a *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" || body.Password == "" {
		http.Error(w, `{"success":false,"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	authResp, err := a.PB.Collection("users").AuthWithPassword(body.Email, body.Password)
	if err != nil {
		http.Error(w, `{"success":false,"error":"authentication failed"}`, http.StatusUnauthorized)
		return
	}

	resp := map[string]any{
		"success": true,
		"token":   authResp.Token,
		"userId":  authResp.Record.Id,
		"email":   authResp.Record.GetString("email"),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// TestConnectionHandler handles GET /test-connection
func (a *App) TestConnectionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := middleware.GetBearerToken(r)
	if err != nil {
		http.Error(w, `{"success":false,"error":"auth required"}`, http.StatusUnauthorized)
		return
	}

	// set token on client
	a.PB.AuthStore.SetToken(token)
	// AuthRefresh to validate & get user record
	auth, err := a.PB.Collection("users").AuthRefresh()
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid token"}`, http.StatusUnauthorized)
		return
	}

	resp := map[string]any{
		"success":   true,
		"message":   "Connection verified successfully",
		"user":      auth.Record.GetString("email"),
		"userId":    auth.Record.Id,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// TestHandler handles POST /challenge01
// multipart/form-data: problemId (int), code=@file
func (a *App) TestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := middleware.GetBearerToken(r)
	if err != nil {
		http.Error(w, `{"success":false,"error":"auth required"}`, http.StatusUnauthorized)
		return
	}

	a.PB.AuthStore.SetToken(token)
	auth, err := a.PB.Collection("users").AuthRefresh()
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid token"}`, http.StatusUnauthorized)
		return
	}
	userID := auth.Record.Id
	userEmail := auth.Record.GetString("email")

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, `{"success":false,"error":"failed to parse form"}`, http.StatusBadRequest)
		return
	}

	pidStr := r.FormValue("problemId")
	pid, convErr := strconv.Atoi(pidStr)
	if convErr != nil || pid <= 0 {
		http.Error(w, `{"success":false,"error":"invalid problemId"}`, http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("code")
	if err != nil {
		http.Error(w, `{"success":false,"error":"missing code file"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	codeBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, `{"success":false,"error":"failed to read code file"}`, http.StatusInternalServerError)
		return
	}

	// Run code in Docker using judge runner
	out, err := judge.RunSubmission(pid, codeBytes)
	if err != nil {
		// Return the runner error
		enc := map[string]any{
			"success": false,
			"error":   fmt.Sprintf("runner error: %s", err.Error()),
		}
		_ = json.NewEncoder(w).Encode(enc)
		return
	}

	// Compare trimmed outputs
	trimmedOut := strings.TrimSpace(out)

	// get expected from config via judge runner's config map (we can call into config)
	expected := "" // safe default; try to fetch via config map if available
	// attempt to import config (avoid circular imports): using SDK is fine; but simplest is to ask judge to return expected,
	// however for this file we will do a small lookup via judge.Config if needed. For clarity set expected via config package:
	// (we'll attempt to fetch via package path)
	// but to avoid circular import complexities we'll assume judge runner wrote a helpful comment: we will fetch using config package directly:
	// (import path)
	//
	// to keep this file compilation-safe, do the following:
	//

	// Fetch expected output via small HTTP-free helper by calling a package-level function.
	// (We assume the package config is accessible at "your-app/internal/config".)
	{
		// lazy import
		// avoid cyc imports: bring in config here:
	}

	// Instead of complicated cross refs, let's get expected using a tiny HTTP-less call to judge's config helper:
	// (we'll call judge.GetExpectedOutput for the problemId; implement it in judge package)

	// Try to get expected via judge package helper (implemented below)
	exp := judge.GetExpectedOutput(pid)

	status := "Rejected"
	if exp != "" && strings.TrimSpace(exp) == trimmedOut {
		status = "Accepted"
	}

	// If accepted: create challengesStages record in PocketBase
	if status == "Accepted" {
		data := map[string]any{
			"user":          userID,
			"challengeName": "redis",
			"completedId":   pid,
		}
		rec, err := a.PB.Collection("challengesStages").Create(data)
		if err != nil {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"error":   fmt.Sprintf("persist failed: %s", err.Error()),
			})
			return
		}
		_ = rec
	}

	resp := map[string]any{
		"success":         status == "Accepted",
		"status":          status,
		"actual_output":   firstN(trimmedOut, 4000),
		"expected_output": firstN(exp, 4000),
		"user_email":      userEmail,
		"timestamp":       time.Now().Format("2006-01-02 15:04:05"),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func firstN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
