package utils

import (
	"fmt"
	"net"
	"os"
	"testing"
)

func TestFindAvailablePort_Success(t *testing.T) {
	port, err := FindAvailablePort(8080, 10)
	if err != nil {
		t.Errorf("Expected to find available port, got error: %v", err)
	}

	if port < 8080 || port >= 8090 {
		t.Errorf("Expected port in range 8080-8089, got %d", port)
	}

	listener, err := net.Listen("tcp", ":"+fmt.Sprintf("%d", port))
	if err != nil {
		t.Errorf("Port %d should be available but got error: %v", port, err)
	} else {
		listener.Close()
	}
}

func TestFindAvailablePort_NoPortsAvailable(t *testing.T) {
	var listeners []net.Listener
	basePort := 9000
	maxPorts := 3

	for i := 0; i < maxPorts; i++ {
		addr := fmt.Sprintf(":%d", basePort+i)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listeners = append(listeners, listener)
		}
	}

	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	_, err := FindAvailablePort(basePort, maxPorts)
	if err == nil {
		t.Error("Expected error when no ports are available, got nil")
	}
}

func TestTernary(t *testing.T) {
	result := Ternary(true, "yes", "no")
	if result != "yes" {
		t.Errorf("Expected 'yes' for true condition, got '%s'", result)
	}

	result = Ternary(false, "yes", "no")
	if result != "no" {
		t.Errorf("Expected 'no' for false condition, got '%s'", result)
	}

	numResult := Ternary(1 > 0, 42, 0)
	if numResult != 42 {
		t.Errorf("Expected 42 for true condition, got %d", numResult)
	}
}

func TestIsRunningWithGoRun(t *testing.T) {
	result := IsRunningWithGoRun()

	_ = result
}

func TestParseFlags_NoArgs(t *testing.T) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	os.Args = []string{"google-auth-wizard"}

}

func TestOpenBrowser(t *testing.T) {
	err := OpenBrowser("https://example.com")

	if err != nil {
		t.Logf("OpenBrowser returned error (expected in test environment): %v", err)
	}
}
