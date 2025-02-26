package update

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"net/http"
	"io"

	"github.com/williamsjokvist/cfn-tracker/pkg/utils"
)

func HandleAutoUpdateTo(version string) error {

	slog.Info(fmt.Sprintf("starting update to version: %s", version))

	zipFileName := getOsSpecificZipFileName()
	downloadLink := fmt.Sprintf("https://github.com/williamsjokvist/cfn-tracker/releases/download/v%s/%s", version, zipFileName)
	//downloadLink := fmt.Sprintf("/home/johan/cfn.zip")
	binaryFileName := getOsSpecificBinaryFileName()

	var zipBytes []byte

	if strings.HasPrefix(downloadLink, "http") {
		res, err := http.Get(downloadLink)
		if err != nil {
			return fmt.Errorf("download latest version %w", err)
		}

		if res.StatusCode != 200 {
			return fmt.Errorf("download latest version, got status: %v", res.Status)
		}

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("read response body: %w", err)
		}
		zipBytes = resBody
	} else {
		bytes, err := os.ReadFile(downloadLink)
		if err != nil {
			return fmt.Errorf("read zip file: %w", err)
		}

		zipBytes = bytes
	}

	// read the whole body
	unzippedFiles, err := utils.UnzipZipFile(zipBytes)
	if err != nil {
		return fmt.Errorf("unzip file: %w", err)
	}

	var exeFileBytes []byte
	for k := range unzippedFiles {
		if strings.HasSuffix(k, binaryFileName) {
			exeFileBytes = unzippedFiles[k]
			break
		}
	}

	if exeFileBytes == nil {
		return fmt.Errorf("find exe in downloaded zip")
	}

	currentExePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get current exe path: %v", err)
	}

	currentExeName := filepath.Base(currentExePath)
	if currentExeName != binaryFileName {
		// This is important to avoid deleting/moving a parent process, like go run, during development/testing
		return fmt.Errorf("current exe name does not match expected name: %s != %s", currentExeName, binaryFileName)
	}

	// Move the current exe to "CFN Tracker.exe.old"
	err = os.Rename(currentExePath, currentExePath+".old")
	if err != nil {
		return fmt.Errorf("rename current exe: %v", err)
	}

	// Write the new exe to the current exe path
	err = os.WriteFile(currentExePath, exeFileBytes, 0755)
	if err != nil {
		return fmt.Errorf("write new exe: %v", err)
	}

	// Launch new process forked
	pid := os.Getpid()
	slog.Info(fmt.Sprintf("launching new process that should know about our pid: %d", pid))
	launchProcessForked(currentExePath, "--auto-update", strconv.Itoa(pid))

	// Exit current process
	slog.Info("exiting current process")
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
