//go:build darwin

package update

import (
	"fmt"
	"os/exec"
	"syscall"
)

func getOsSpecificZipFileName() string {
	if isArmCpu() {
		return "cfn-tracker-darwin-arm64.zip"
	} else {
		return "cfn-tracker-darwin-amd64.zip"
	}
}

func getOsSpecificBinaryFileName() string {
	return "CFN Tracker"
}

func launchProcessForked(binaryFilePath string, args ...string) {
	cmd := exec.Command(binaryFilePath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	err := cmd.Start()
	if err != nil {
		panic(fmt.Errorf("start new process: %w", err))
	}
}
