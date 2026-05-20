package versioncheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultURL = "https://api.github.com/repos/MagnumGOYB/memforge/releases/latest"

type Result struct {
	Latest  string
	Current string
}

func Check(ctx context.Context, storageRoot string, current string) (Result, error) {
	if latest := strings.TrimSpace(os.Getenv("MEMFORGE_VERSION_CHECK_LATEST")); latest != "" {
		latest = normalizeVersion(latest)
		writeCache(filepath.Join(storageRoot, "cache", "version-check.json"), latest)
		return compare(latest, current), nil
	}
	url := strings.TrimSpace(os.Getenv("MEMFORGE_VERSION_CHECK_URL"))
	if url == "" {
		url = defaultURL
	}
	cachePath := filepath.Join(storageRoot, "cache", "version-check.json")
	if cached, ok := readFreshCache(cachePath); ok {
		return compare(cached, current), nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}
	client := http.Client{Timeout: 750 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("version check returned %s", resp.Status)
	}
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Result{}, err
	}
	latest := normalizeVersion(payload.TagName)
	if latest == "" {
		return Result{}, fmt.Errorf("version check response missing tag_name")
	}
	writeCache(cachePath, latest)
	return compare(latest, current), nil
}

func compare(latest string, current string) Result {
	return Result{Latest: normalizeVersion(latest), Current: normalizeVersion(current)}
}

func (r Result) HasUpdate() bool {
	return r.Latest != "" && r.Current != "" && r.Latest != r.Current && r.Current != "dev"
}

func readFreshCache(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var payload struct {
		Latest    string    `json:"latest"`
		CheckedAt time.Time `json:"checked_at"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", false
	}
	if payload.Latest == "" || time.Since(payload.CheckedAt) > 24*time.Hour {
		return "", false
	}
	return payload.Latest, true
}

func writeCache(path string, latest string) {
	payload := struct {
		Latest    string    `json:"latest"`
		CheckedAt time.Time `json:"checked_at"`
	}{Latest: latest, CheckedAt: time.Now().UTC()}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(path, append(data, '\n'), 0o644)
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}
