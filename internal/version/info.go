package version

import (
	"fmt"
	"os"
)

var (
	// Version is injected during build time
	Version = "dev"
	// BuildTime is injected during build time
	BuildTime = "unknown"
)

// PrintBanner prints the application banner with version information and exits
func PrintBanner(appName string) {
	fmt.Printf("=========================================\n")
	fmt.Printf("  PostQ-Tunnel\n")
	fmt.Printf("  %s\n", appName)
	fmt.Printf("  Version: %s\n", Version)
	fmt.Printf("=========================================\n")
	os.Exit(0)
}
