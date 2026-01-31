package license

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"

	"github.com/denisbrodbeck/machineid"
)

func GetHWID() (string, error) {
	return machineid.ProtectedID("ConsoleSniper")
}

func getDiskSerial() string {
	// Expert Level: Get physical disk serial via WMIC (harder to spoof than partition ID)
	out, err := exec.Command("wmic", "diskdrive", "get", "serialnumber").Output()
	if err != nil {
		return "no-disk-serial"
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1])
	}
	return "unknown-disk"
}

func getMACAddress() string {
	// Expert Level: Get MAC of the first non-virtual physical adapter
	interfaces, err := net.Interfaces()
	if err != nil {
		return "no-mac"
	}
	for _, i := range interfaces {
		if i.Flags&net.FlagLoopback == 0 && i.HardwareAddr != nil {
			return i.HardwareAddr.String()
		}
	}
	return "unknown-mac"
}

func GetEntropicHWID() (string, error) {
	// Vector 1: Machine ID (Standard)
	mid, _ := machineid.ID()

	// Vector 2: Physical Hardware (Deep)
	disk := getDiskSerial()
	mac := getMACAddress()

	// Vector 3: Environment Context
	arch := runtime.GOARCH
	os := runtime.GOOS
	cpus := runtime.NumCPU()

	// Combine all vectors into a unique fingerprint
	raw := fmt.Sprintf("v2-%s-%s-%s-%s-%s-%d-ConsoleSniper-Hardcore", mid, disk, mac, arch, os, cpus)
	
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", hash), nil
}