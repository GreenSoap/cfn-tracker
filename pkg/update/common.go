package update

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-version"
	"github.com/williamsjokvist/cfn-tracker/pkg/browser"
	"github.com/williamsjokvist/cfn-tracker/pkg/utils"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Prob inject this
var restyClient = resty.New()

func HandleAutoUpdateTo(latestVersion string) error {

	slog.Info(fmt.Sprintf(`HandleAutoUpdate: Starting update to version: %s`, latestVersion))

	zipFileName := getOsSpecificZipFileName()
	downloadLink := fmt.Sprintf("https://github.com/williamsjokvist/cfn-tracker/releases/download/v%s/%s", latestVersion, zipFileName)
	//downloadLink := fmt.Sprintf("/home/johan/cfn.zip")
	binaryFileName := getOsSpecificBinaryFileName()

	request := restyClient.R()

	var zipBytes []byte

	if strings.HasPrefix(downloadLink, "http") {

		res, err := request.Get(downloadLink)
		if err != nil {
			return fmt.Errorf(`HandleAutoUpdate: Failed to download latest version: %v: %s`, err, downloadLink)
		}

		if res.StatusCode() != 200 {
			return fmt.Errorf(`HandleAutoUpdate: Failed to download latest version: %v: %s`, res.Status(), downloadLink)
		}

		zipBytes = res.Body()
	} else {
		bytes, err := os.ReadFile(downloadLink)
		if err != nil {
			return fmt.Errorf(`HandleAutoUpdate: Failed to read zip file: %v`, err)
		}

		zipBytes = bytes
	}

	// read the whole body
	unzippedFiles, err := utils.UnzipZipFile(zipBytes)
	if err != nil {
		return fmt.Errorf(`HandleAutoUpdate: Failed to unzip downloaded zip: %v`, err)
	}

	var exeFileBytes []byte
	for k := range unzippedFiles {
		if strings.HasSuffix(k, binaryFileName) {
			exeFileBytes = unzippedFiles[k]
			break
		}
	}

	if exeFileBytes == nil {
		return fmt.Errorf(`HandleAutoUpdate: Failed to find exe in downloaded zip`)
	}

	currentExePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf(`HandleAutoUpdate: Failed to get current exe path: %v`, err)
	}

	currentExeName := filepath.Base(currentExePath)
	if currentExeName != binaryFileName {
		// This is important to avoid deleting/moving a parent process, like go run, during development/testing
		return fmt.Errorf(`HandleAutoUpdate: Current exe name does not match expected name: %s != %s`, currentExeName, binaryFileName)
	}

	// Move the current exe to "CFN Tracker.exe.old"
	err = os.Rename(currentExePath, currentExePath+`.old`)
	if err != nil {
		return fmt.Errorf(`HandleAutoUpdate: Failed to rename current exe: %v`, err)
	}

	// Write the new exe to the current exe path
	err = os.WriteFile(currentExePath, exeFileBytes, 0755)
	if err != nil {
		return fmt.Errorf(`HandleAutoUpdate: Failed to write new exe: %v`, err)
	}

	// Launch new process forked
	pid := os.Getpid()
	slog.Info(fmt.Sprintf(`HandleAutoUpdate: Launching new process that should know about our pid: %d`, pid))
	launchProcessForked(currentExePath, "--auto-update", strconv.Itoa(pid))

	// Exit current process
	slog.Info(`HandleAutoUpdate: Exiting current process`)
	os.Exit(0)

	return nil
}

func isArmCpu() bool {
	switch runtime.GOARCH {
	case "arm", "arm64":
		return true
	default:
		return false
	}
}

func GetVersions(currentVersionStr string, b *browser.Browser) (*version.Version, *version.Version, error) {
	currentVersion, err := version.NewVersion(currentVersionStr)
	if err != nil {
		log.Println(err)
		return nil, nil, fmt.Errorf(`failed to parse current app version: %w`, err)
	}
	latestVersion, err := b.GetLatestAppVersion()
	if err != nil {
		log.Println(err)
		return nil, nil, fmt.Errorf(`failed to check for update: %w`, err)
	}

	return currentVersion, latestVersion, nil
}

func HandleAutoUpdate(currentVersionStr string, b *browser.Browser) error {

	currentVersion, latestVersion, err := GetVersions(currentVersionStr, b)
	if err != nil {
		return err
	}

	hasUpdate := currentVersion.LessThan(latestVersion)
	slog.Info("HandleAutoUpdate", "Has update", hasUpdate, "Current", currentVersion.String(), "Latest", latestVersion.String())

	if !hasUpdate {
		return nil
	}

	return HandleAutoUpdateTo(latestVersion.String())
}
