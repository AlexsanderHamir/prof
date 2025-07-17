package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	CurrentVersion = "1.0.30"
	GitHubRepo     = "AlexsanderHamir/prof"
	GitHubAPIURL   = "https://api.github.com/repos/AlexsanderHamir/prof/releases/latest"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}

func getLatestVersion() (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(GitHubAPIURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

func Check() (string, string) {
	latest, err := getLatestVersion()
	if err != nil {
		return CurrentVersion, ""
	}
	return CurrentVersion, latest
}

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
