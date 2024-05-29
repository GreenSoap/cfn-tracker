//go:build windows

package update

import (
	"fmt"
	"os/exec"
	"syscall"
)

func getOsSpecificZipFileName() string {
	if isArmCpu() {
		return "cfn-tracker-windows-arm64.zip"
	} else {
		return "cfn-tracker-windows-amd64.zip"
	}
}

func getOsSpecificBinaryFileName() string {
	return "CFN Tracker.exe"
}

func launchProcessForked(binaryFilePath string, args ...string) {
	cmd := exec.Command(binaryFilePath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	err := cmd.Start()
	if err != nil {
		panic(fmt.Sprintf(`failed to start new process: %v`, err))
	}
}
