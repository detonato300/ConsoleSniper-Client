package main

import (
	"agent_client/internal/api"
	"agent_client/internal/license"
	"agent_client/internal/security"
	"agent_client/internal/ui"
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const Version = "v3.4.5"
const DevSecret = "DEVSNIPER2026"

func main() {
	// Parse flags for Debug mode
	debugEnabled := false
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--DEBUG") {
			code := strings.TrimPrefix(arg, "--DEBUG")
			if code == DevSecret {
				debugEnabled = true
			}
		}
	}

	if !debugEnabled {
		security.PerformSecurityAudit()
	}

	// Check for environment override, otherwise use production VPS
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://consolesniper.aethernatolith.run.place"
	}

	token := os.Getenv("VOLUNTEER_TOKEN")
	if token == "" {
		token = "default-dev-token"
	}

	hwid, _ := license.GetEntropicHWID()
	client := api.NewClient(baseURL, token, hwid)
	
	m := ui.InitialModel(client)
	m.DebugEnabled = debugEnabled
	
	// Load local license
	licData, err := license.LoadLicenseEncrypted("license.key", hwid)
	if err == nil {
		// New Public Key XOR 0x42
		obfKey := []byte{0x7b, 0x7b, 0x26, 0x77, 0x70, 0x75, 0x26, 0x73, 0x27, 0x7a, 0x26, 0x72, 0x73, 0x26, 0x7a, 0x71, 0x72, 0x75, 0x73, 0x76, 0x74, 0x74, 0x21, 0x75, 0x27, 0x26, 0x75, 0x71, 0x7b, 0x7a, 0x24, 0x7b, 0x27, 0x21, 0x20, 0x73, 0x73, 0x72, 0x7a, 0x70, 0x71, 0x26, 0x72, 0x7b, 0x72, 0x76, 0x27, 0x26, 0x75, 0x24, 0x7b, 0x74, 0x27, 0x74, 0x77, 0x27, 0x75, 0x71, 0x71, 0x27, 0x7a, 0x20, 0x74, 0x70}
		pubKey := security.Deobfuscate(obfKey, 0x42)
		
		payload, vErr := license.VerifyLicense(licData, pubKey)
		if vErr == nil {
			m.Unlicensed = false
			m.LicenseData = payload
			m.Status = "Welcome back!"
            // CRITICAL FIX: Update API client with the loaded license token
            client = api.NewClient(baseURL, string(licData), hwid)
            m.Client = client
		}
	}

	status, minVer, err := client.CheckVersion(context.Background(), Version)
	if err == nil {
		m.VersionStatus = status
		m.MinVersion = minVer
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}