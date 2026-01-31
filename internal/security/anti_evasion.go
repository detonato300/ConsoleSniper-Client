package security

import (
	"net"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

var blacklistedProcesses = []string{
	"cheatengine",
	"processhacker",
	"x64dbg",
	"fiddler",
	"wireshark",
	"httpdebugger",
	"charles",
}

func isBlacklisted(name string) bool {
	name = strings.ToLower(name)
	for _, b := range blacklistedProcesses {
		if strings.Contains(name, b) {
			return true
		}
	}
	return false
}

// ScanProcesses checks for blacklisted tools.
func ScanProcesses() bool {
	procs, err := process.Processes()
	if err != nil {
		return false
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		if isBlacklisted(name) {
			return true
		}
	}
	return false
}

// PerformSecurityAudit runs all security checks and flags the environment if compromised.
func PerformSecurityAudit() {
	if CheckDebugger() || CheckVM() || ScanProcesses() {
		GlobalState.MarkTainted()
	}
	
	// Add Jitter check (threshold: 2000 cycles)
	jitter := MeasureJitter()
	if IsHighJitter(jitter, 2000) {
		GlobalState.MarkTainted()
	}
}

// CheckDebugger returns true if a debugger is attached to the process.
func CheckDebugger() bool {
	return CheckDebuggerWin()
}

// CheckCPUID uses assembly level checks to detect hypervisor presence.
func CheckCPUID() bool {
	// Standard CPUID call with EAX=1
	// If 31st bit of ECX is 1, it's a VM
	return cpuid_check()
}

// CheckHardwareResources checks for suspiciously low hardware specs common in VMs.
func CheckHardwareResources() bool {
	// Vector 1: Low CPU core count
	if runtime.NumCPU() < 2 {
		return true
	}
	return false
}

// CheckVM returns true if the environment looks like a Virtual Machine.
func CheckVM() bool {
	// Vector 1: CPUID Hypervisor Bit (Highest Reliability)
	if CheckCPUID() {
		return true
	}

	// Vector 2: Hardware Resources
	if CheckHardwareResources() {
		return true
	}

	// Vector 3: MAC Address Prefix check
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			addr := i.HardwareAddr.String()
			// Common VM MAC prefixes (VMware, VirtualBox, Hyper-V)
			prefixes := []string{"08:00:27", "00:05:69", "00:0C:29", "00:50:56", "00:15:5D"}
			for _, p := range prefixes {
				if strings.HasPrefix(strings.ToUpper(addr), p) {
					return true
				}
			}
		}
	}

	return CheckVMFilesWin()
}
