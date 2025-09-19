package utils

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func IsRunningWithGoRun() bool {
	execPath := os.Args[0]

	return strings.Contains(execPath, os.TempDir()) ||
		strings.Contains(execPath, "go-build")
}

func ParseFlags() string {
	var filename string
	var forceNew bool
	var clearTokens bool

	flag.StringVar(&filename, "file", "", "Path to JSON file")
	flag.StringVar(&filename, "f", "", "Path to JSON file (shortcut)")
	flag.BoolVar(&forceNew, "force-new", false, "Force getting a new token (ignore saved tokens)")
	flag.BoolVar(&forceNew, "n", false, "Force getting a new token (shortcut)")
	flag.BoolVar(&clearTokens, "clear-tokens", false, "Clear all saved tokens and exit")
	flag.BoolVar(&clearTokens, "c", false, "Clear all saved tokens and exit (shortcut)")
	flag.Parse()

	if clearTokens {
		clearSavedTokens()
		os.Exit(0)
	}

	if filename == "" {
		printUsage()
		os.Exit(1)
	}

	if forceNew {
		os.Setenv("GOOGLE_AUTH_WIZARD_FORCE_NEW", "true")
	}

	return filename
}

func printUsage() {
	fmt.Println("Error: No file specified")
	fmt.Println("\nUsage:")
	fmt.Printf("  %s -file <client_secret_[id].apps.googleusercontent.com.json>\n", Ternary(IsRunningWithGoRun(), "go run main.go", "./google-auth-wizard"))
	fmt.Printf("  %s -f <client_secret_[id].apps.googleusercontent.com.json>\n", Ternary(IsRunningWithGoRun(), "go run main.go", "./google-auth-wizard"))
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Printf("  %s -f credentials.json                 # Use saved token if available\n", Ternary(IsRunningWithGoRun(), "go run main.go", "./google-auth-wizard"))
	fmt.Printf("  %s -f credentials.json -n              # Force new token\n", Ternary(IsRunningWithGoRun(), "go run main.go", "./google-auth-wizard"))
	fmt.Printf("  %s -c                                  # Clear saved tokens\n", Ternary(IsRunningWithGoRun(), "go run main.go", "./google-auth-wizard"))
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  GOOGLE_AUTH_WIZARD_DEBUG=true         # Enable debug logging")
	fmt.Println("  GOOGLE_AUTH_WIZARD_VERBOSE=true       # Enable verbose logging")
	fmt.Println("  GOOGLE_AUTH_WIZARD_SILENT=true        # Silent mode")
}

func ReadCredentials(filename string) []byte {
	credentials, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to read client secret file: %w", err))
	}
	return credentials
}

func FindAvailablePort(defaultPort int, maxPortTries int) (int, error) {
	for port := defaultPort; port < defaultPort+maxPortTries; port++ {
		if isPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found in range %d-%d", defaultPort, defaultPort+maxPortTries-1)
}

func isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func Ternary[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func clearSavedTokens() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	tokenPath := filepath.Join(homeDir, ".google-auth-wizard", "token.json")

	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		fmt.Println("No saved tokens found.")
		return
	}

	if err := os.Remove(tokenPath); err != nil {
		fmt.Printf("Error removing token file: %v\n", err)
		return
	}

	fmt.Println("Saved tokens cleared successfully.")
}
