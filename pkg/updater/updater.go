package updater

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Yehya-Elsawy/explain/pkg/ui"
)

const Repo = "Yehya-Elsawy/explain"

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
}

type githubTag struct {
	Name string `json:"name"`
}

// FetchLatestVersion gets the latest release or tag from GitHub.
func FetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	// 1. Try releases/latest first
	req, err := http.NewRequest("GET", "https://api.github.com/repos/"+Repo+"/releases/latest", nil)
	if err == nil {
		req.Header.Set("User-Agent", "explain-cli-updater")
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var rel githubRelease
				if err := json.NewDecoder(resp.Body).Decode(&rel); err == nil && rel.TagName != "" {
					return rel.TagName, nil
				}
			}
		}
	}

	// 2. Fallback to /tags
	reqTags, err := http.NewRequest("GET", "https://api.github.com/repos/"+Repo+"/tags", nil)
	if err != nil {
		return "", err
	}
	reqTags.Header.Set("User-Agent", "explain-cli-updater")
	respTags, err := client.Do(reqTags)
	if err != nil {
		return "", err
	}
	defer respTags.Body.Close()

	if respTags.StatusCode == http.StatusOK {
		var tags []githubTag
		if err := json.NewDecoder(respTags.Body).Decode(&tags); err == nil && len(tags) > 0 {
			return tags[0].Name, nil
		}
	}

	return "", fmt.Errorf("no releases or tags found on GitHub repository (%s)", Repo)
}

// SelfUpdate downloads and replaces the current binary with the latest release.
func SelfUpdate(currentVersion string) error {
	fmt.Println()
	fmt.Printf("  %s %s\n", ui.Colorize(ui.BoldYellow, "[>]"), ui.Colorize(ui.BoldWhite, "Checking for updates on GitHub..."))

	latestTag, err := FetchLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	cleanCurrent := strings.TrimPrefix(currentVersion, "v")
	cleanLatest := strings.TrimPrefix(latestTag, "v")

	if (cleanCurrent == cleanLatest && cleanCurrent != "") || cleanCurrent >= cleanLatest {
		fmt.Printf("  %s %s %s\n\n", ui.Colorize(ui.BoldGreen, "✓"), ui.Colorize(ui.BoldGreen, "explain is already up to date"), ui.Colorize(ui.Dim, "("+currentVersion+")"))
		return nil
	}

	fmt.Printf("  %s %s %s %s\n",
		ui.Colorize(ui.BoldYellow, "[>]"),
		ui.Colorize(ui.BoldWhite, "Found new version:"),
		ui.Colorize(ui.BoldGreen, latestTag),
		ui.Colorize(ui.Dim, "(current: "+currentVersion+")"),
	)

	osName := runtime.GOOS
	arch := runtime.GOARCH
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/explain_%s_%s.tar.gz", Repo, latestTag, osName, arch)

	fmt.Printf("  %s %s\n", ui.Colorize(ui.BoldYellow, "[>]"), ui.Colorize(ui.White, "Downloading binary for "+osName+"/"+arch+"..."))
	showProgressBar()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with HTTP %d from %s", resp.StatusCode, downloadURL)
	}

	// Extract binary from tar.gz in-memory
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to decompress gzip archive: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var binaryBytes []byte

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar archive: %w", err)
		}
		if header.Name == "explain" || strings.HasSuffix(header.Name, "/explain") {
			binaryBytes, err = io.ReadAll(tr)
			if err != nil {
				return fmt.Errorf("failed to read binary from archive: %w", err)
			}
			break
		}
	}

	if len(binaryBytes) == 0 {
		return fmt.Errorf("could not find 'explain' binary inside release archive")
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine current binary path: %w", err)
	}

	// Write to temporary file next to binary, then atomically rename
	tmpFile := execPath + ".tmp"
	if err := os.WriteFile(tmpFile, binaryBytes, 0755); err != nil {
		return fmt.Errorf("permission denied writing to %s (try running with sudo explain update): %w", execPath, err)
	}

	if err := os.Rename(tmpFile, execPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to replace executable: %w", err)
	}

	fmt.Printf("\n  %s %s\n", ui.Colorize(ui.BoldGreen, "✓"), ui.Colorize(ui.BoldGreen, "Successfully updated explain to "+latestTag+"!"))
	fmt.Printf("    %s %s\n\n", ui.Colorize(ui.Dim, "Binary path:"), ui.Colorize(ui.White, execPath))
	return nil
}

func showProgressBar() {
	width := 30
	for i := 1; i <= width; i++ {
		filled := strings.Repeat("█", i)
		empty := strings.Repeat("░", width-i)
		percent := (i * 100) / width
		fmt.Printf("\r    %s%s%s%s %d%%", ui.Cyan, "["+filled, empty+"]", ui.Reset, percent)
		time.Sleep(12 * time.Millisecond)
	}
	fmt.Printf("\r    %s[██████████████████████████████]%s 100%%\n", ui.Green, ui.Reset)
}
