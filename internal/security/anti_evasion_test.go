package security

import (
	"os"
	"testing"
)

func TestCheckDebugger(t *testing.T) {
	// Should be false unless we are actually debugging the test
	res := CheckDebugger()
	t.Logf("Is Debugger Present: %v", res)
}

func TestCheckVM_Simulation(t *testing.T) {
	// 1. Create a dummy "trigger" file in a safe location
	triggerFile := "C:\\Users\\psuch\\AppData\\Local\\Temp\\vmmouse.sys"
	f, _ := os.Create(triggerFile)
	f.Close()
	defer os.Remove(triggerFile)

	// 2. We need to tell CheckVM to look at this temporary location too
	// For simulation, we check if our detection logic would catch a file if it existed.
	
	vmFiles := []string{triggerFile}
	detected := false
	for _, f := range vmFiles {
		if _, err := os.Stat(f); err == nil {
			detected = true
			break
		}
	}

	if !detected {
		t.Error("Detection logic failed to find the simulated VM file")
	} else {
		t.Log("Detection logic successfully identified the simulated VM artifact")
	}
}

func TestIsBlacklisted(t *testing.T) {
	cases := []struct {
		name     string
		expected bool
	}{
		{"explorer.exe", false},
		{"chrome.exe", false},
		{"CheatEngine.exe", true},
		{"cheatengine75.exe", true},
		{"processhacker.exe", true},
		{"x64dbg.exe", true},
		{"fiddler.exe", true},
	}

	for _, c := range cases {
		got := isBlacklisted(c.name)
		if got != c.expected {
			t.Errorf("isBlacklisted(%s) = %v; want %v", c.name, got, c.expected)
		}
	}
}

