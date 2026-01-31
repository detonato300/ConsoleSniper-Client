package api

import (
	"agent_client/internal/security"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID      int                    `json:"id"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type Client struct {
	BaseURL        string
	VolunteerToken string
	HWID           string
	SessionKey     []byte
}

func shouldSimulateFailure() bool {
	if security.GlobalState.IsTainted() {
		// 10% chance to simulate a random failure
		return rand.Intn(10) == 0
	}
	return false
}

func NewClient(baseURL, token, hwid string) *Client {
	// Derive User Secret first (HMAC of Master Secret + Token)
	// "top-secret-master-key" XOR 0x42
	obfMaster := []byte{0x36, 0x2d, 0x32, 0x6f, 0x31, 0x27, 0x21, 0x30, 0x27, 0x36, 0x6f, 0x2f, 0x23, 0x31, 0x36, 0x27, 0x30, 0x6f, 0x29, 0x27, 0x3b}
	master := security.Deobfuscate(obfMaster, 0x42)

	h := hmac.New(sha256.New, []byte(master))
	h.Write([]byte(token))
	userSecret := hex.EncodeToString(h.Sum(nil))

	// Session salt - in prod would be exchanged
	sessKey, _ := security.DeriveSessionKey(userSecret, "cs-session-v1", "api-request-signing")

	return &Client{
		BaseURL:        baseURL,
		VolunteerToken: token,
		HWID:           hwid,
		SessionKey:     sessKey,
	}
}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func (c *Client) signRequest(req *http.Request, body []byte) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := uuid.New().String()

	// 1. Header Signature
	hMsg := fmt.Sprintf("%s:%s:%s", ts, c.VolunteerToken, nonce)
	h := hmac.New(sha256.New, c.SessionKey)
	h.Write([]byte(hMsg))
	hSig := hex.EncodeToString(h.Sum(nil))

	// 2. Body Signature
	bMsg := fmt.Sprintf("%s:%s", ts, string(body))
	if len(body) == 0 {
		bMsg = fmt.Sprintf("%s:{}", ts)
	}
	b := hmac.New(sha256.New, c.SessionKey)
	b.Write([]byte(bMsg))
	bSig := hex.EncodeToString(b.Sum(nil))

	req.Header.Set("X-Timestamp", ts)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-License-Sig", c.VolunteerToken)
	req.Header.Set("X-HWID", c.HWID)
	req.Header.Set("X-Header-Signature", hSig)
	req.Header.Set("X-Signature", bSig)
}

func (c *Client) ReadyForTask(ctx context.Context, capabilities []string, aiStatus string, aiQuota int, currentVersion string) (*Task, error) {
	if shouldSimulateFailure() {
		return nil, fmt.Errorf("security trigger: simulated 500 (client tainted)")
	}

	payload := map[string]interface{}{
		"capabilities":    capabilities,
		"ai_status":       aiStatus,
		"ai_quota":        aiQuota,
		"current_version": currentVersion,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	// 1. End-to-End Encryption
	encBody, err := security.Encrypt(body, c.SessionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/v1/tasks/ready", bytes.NewBuffer(encBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.signRequest(req, encBody)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Encrypted", "true")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 2. Decrypt Response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	decBody, err := security.Decrypt(respBody, c.SessionKey)
	if err != nil {
		// Fallback: maybe server didn't encrypt?
		decBody = respBody
	}

	var result struct {
		Task   *Task                  `json:"task"`
		Config map[string]interface{} `json:"config"`
	}
	if err := json.Unmarshal(decBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Update local AI configuration if provided by server
	if result.Config != nil {
		if geminiKey, ok := result.Config["gemini_api_key"].(string); ok && geminiKey != "" {
			// Persist synchronized key to local storage
			_ = SaveAIConfig(geminiKey)
		}
	}

	return result.Task, nil
}

func (c *Client) SubmitTask(ctx context.Context, taskID int, resultData interface{}, status string, metadata interface{}) (map[string]interface{}, error) {
	if shouldSimulateFailure() {
		return nil, fmt.Errorf("security trigger: simulated 500 (client tainted)")
	}

	payload := map[string]interface{}{
		"task_id":  taskID,
		"result":   resultData,
		"status":   status,
		"metadata": metadata,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	// 1. End-to-End Encryption
	encBody, err := security.Encrypt(body, c.SessionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/v1/tasks/submit", bytes.NewBuffer(encBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.signRequest(req, encBody)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Encrypted", "true")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 2. Decrypt Response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	decBody, err := security.Decrypt(respBody, c.SessionKey)
	if err != nil {
		decBody = respBody
	}

	var res struct {
		Gamification map[string]interface{} `json:"gamification"`
	}
	if err := json.Unmarshal(decBody, &res); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return res.Gamification, nil
}

func (c *Client) GetStats(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/api/v1/user/stats", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.signRequest(req, nil)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch stats: %d", resp.StatusCode)
	}

	// Decrypt Response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	decBody, err := security.Decrypt(respBody, c.SessionKey)
	if err != nil {
		decBody = respBody
	}

	var result map[string]interface{}
	if err := json.Unmarshal(decBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return result, nil
}

func (c *Client) CheckVersion(ctx context.Context, currentVersion string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/version/check?version=%s", c.BaseURL, currentVersion), nil)
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status     string `json:"status"`
		MinVersion string `json:"min_version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}

	return result.Status, result.MinVersion, nil
}

type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func GetLatestRelease(ctx context.Context) (*ReleaseInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/detonato300/ConsoleSniper-Client/releases/latest", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &release, nil
}

func (c *Client) RedeemPoints(ctx context.Context, rewardType string) error {
	payload := map[string]string{
		"reward_type": rewardType,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	// End-to-End Encryption
	encBody, err := security.Encrypt(body, c.SessionKey)
	if err != nil {
		return fmt.Errorf("encrypt body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/v1/user/redeem", bytes.NewBuffer(encBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	c.signRequest(req, encBody)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Encrypted", "true")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("redemption failed: %d", resp.StatusCode)
	}

	return nil
}
