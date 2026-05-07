package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cristiangonsevi/orbit/internal/ui"
	"github.com/cristiangonsevi/orbit/internal/utils"
	"github.com/spf13/cobra"
)

const (
	repoOwner   = "cristiangonsevi"
	repoName    = "orbit"
	githubAPI   = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases"
	githubDL    = "https://github.com/" + repoOwner + "/" + repoName + "/releases/download"
)

var (
	updateCheck   bool
	updateYes     bool
	updateVersion string
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func parseVersion(version string) (string, error) {
	re := regexp.MustCompile(`^v?(\d+\.\d+\.\d+)$`)
	matches := re.FindStringSubmatch(version)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid version format: %s", version)
	}
	if strings.HasPrefix(version, "v") {
		return matches[1], nil
	}
	return matches[1], nil
}

func compareVersions(current, latest string) (int, error) {
	currParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	for i := 0; i < 3; i++ {
		var c, l int
		fmt.Sscanf(currParts[i], "%d", &c)
		fmt.Sscanf(latestParts[i], "%d", &l)
		if c < l {
			return -1, nil
		}
		if c > l {
			return 1, nil
		}
	}
	return 0, nil
}

func getLatestReleaseTag() (string, error) {
	resp, err := http.Get(githubAPI + "/latest")
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to parse release JSON: %w", err)
	}

	cleanVersion, err := parseVersion(release.TagName)
	if err != nil {
		return "", fmt.Errorf("failed to parse tag name: %w", err)
	}

	return cleanVersion, nil
}

func getExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	realpath, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}
	return realpath, nil
}

func downloadBinary(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read download: %w", err)
	}

	if err := os.WriteFile(destPath, data, 0755); err != nil {
		return fmt.Errorf("failed to write binary: %w", err)
	}

	return nil
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update orbit to the latest version",
	Long: `Check for updates and update orbit to the latest version.

Use --check to only verify the current version without installing.
Use --yes to skip the confirmation prompt.
Use --version to install a specific version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		platform, err := utils.DetectPlatform()
		if err != nil {
			ui.Error(err.Error())
			return err
		}

		currentVersion := Version
		targetVersion := updateVersion

		if targetVersion == "" {
			targetVersion, err = getLatestReleaseTag()
			if err != nil {
				ui.Warning(fmt.Sprintf("Could not check for updates: %v", err))
				return err
			}
		} else {
			targetVersion, err = parseVersion(targetVersion)
			if err != nil {
				ui.Error(fmt.Sprintf("Invalid version format: %v", err))
				return err
			}
		}

		if updateCheck {
			if targetVersion == "" {
				ui.Info("Could not determine latest version")
			} else {
				cmp, err := compareVersions(currentVersion, targetVersion)
				if err != nil {
					ui.Error(err.Error())
					return err
				}
				if cmp < 0 {
					ui.Warning(fmt.Sprintf("A new version is available: %s → %s", currentVersion, targetVersion))
				} else {
					ui.Success(fmt.Sprintf("You're up to date! (%s)", currentVersion))
				}
			}
			return nil
		}

		cmp, err := compareVersions(currentVersion, targetVersion)
		if err != nil {
			ui.Error(err.Error())
			return err
		}

		if cmp >= 0 && updateVersion == "" {
			ui.Success(fmt.Sprintf("You're up to date! (%s)", currentVersion))
			return nil
		}

		if !updateYes {
			ui.Warning(fmt.Sprintf("A new version is available: %s → %s", currentVersion, targetVersion))
			fmt.Print("Do you want to update? [y/N] ")
			var confirm string
			fmt.Scan(&confirm)
			if confirm != "y" && confirm != "Y" {
				ui.Info("Update cancelled")
				return nil
			}
		}

		ui.Info(fmt.Sprintf("Updating to version %s...", targetVersion))

		exePath, err := getExecutablePath()
		if err != nil {
			ui.Error(err.Error())
			return err
		}

		tag := targetVersion
		if !strings.HasPrefix(tag, "v") {
			tag = "v" + tag
		}

		downloadURL := fmt.Sprintf("%s/%s/orbit-%s", githubDL, tag, platform)

		tmpDir, err := os.MkdirTemp("", "orbit-update")
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to create temp dir: %v", err))
			return err
		}
		defer os.RemoveAll(tmpDir)

		tmpBinary := filepath.Join(tmpDir, "orbit-"+platform.String())

		if err := downloadBinary(downloadURL, tmpBinary); err != nil {
			ui.Error(fmt.Sprintf("Failed to download: %v", err))
			return err
		}

		backupPath := filepath.Join("/tmp", fmt.Sprintf("orbit-backup-%s-%s", currentVersion, platform.String()))

		if err := exec.Command("cp", exePath, backupPath).Run(); err != nil {
			ui.Warning(fmt.Sprintf("Failed to create backup: %v", err))
		} else {
			ui.Info(fmt.Sprintf("Backup saved to %s", backupPath))
		}

		if err := exec.Command("cp", tmpBinary, exePath).Run(); err != nil {
			ui.Error(fmt.Sprintf("Failed to install: %v", err))
			return fmt.Errorf("installation failed")
		}

		ui.Success(fmt.Sprintf("Updated to version %s", targetVersion))

		checkVersionOutput := &bytes.Buffer{}
		execCmd := exec.Command(exePath, "version")
		execCmd.Stdout = checkVersionOutput
		execCmd.Stderr = os.Stderr
		if err := execCmd.Run(); err == nil {
			ui.Info(fmt.Sprintf("New version: %s", strings.TrimSpace(checkVersionOutput.String())))
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Check for updates without installing")
	updateCmd.Flags().BoolVar(&updateYes, "yes", false, "Skip confirmation prompt")
	updateCmd.Flags().StringVar(&updateVersion, "version", "", "Install a specific version")
	rootCmd.AddCommand(updateCmd)
}
