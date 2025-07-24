package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CurrentVersion is the current version of the prof tool.
const (
	currentVersion = "1.0.4"
	gitHubAPIURL   = "https://api.github.com/repos/AlexsanderHamir/prof/releases/latest"
	waitTime       = 10
)

// GitHubRelease represents the structure of a GitHub release response.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// normalizeVersion removes the 'v' prefix from a version string.
func normalizeVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}

// getLatestVersion fetches the latest release tag from GitHub.
func getLatestVersion(ctx context.Context) (tagName string, err error) {
	client := &http.Client{
		Timeout: waitTime * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gitHubAPIURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest version: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("response body close failed: %w", closeErr)
			}
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode GitHub release: %w", err)
	}

	return release.TagName, nil
}

// Check returns the current and latest available version.
func Check() (string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), waitTime*time.Second)
	defer cancel()

	latest, err := getLatestVersion(ctx)
	if err != nil {
		return currentVersion, ""
	}
	return currentVersion, latest
}

// FormatOutput formats the version information for display.
func FormatOutput(current, latest string) string {
	output := fmt.Sprintf("Current version: %s\n", current)

	if latest == "" {
		output += "Latest version: Unable to fetch (check your internet connection)\n"
		return output
	}

	normalizedCurrent := normalizeVersion(current)
	normalizedLatest := normalizeVersion(latest)

	if normalizedLatest == normalizedCurrent {
		output += fmt.Sprintf("Latest version: %s (up to date)\n", latest)
	} else {
		output += fmt.Sprintf("Latest version: %s (update available)\n", latest)
	}

	return output
}
