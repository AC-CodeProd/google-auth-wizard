package logger

import (
	"os"
	"testing"
)

func TestLogLevels(t *testing.T) {
	if LEVEL_SILENT >= LEVEL_ERROR {
		t.Error("LEVEL_SILENT should be less than LEVEL_ERROR")
	}

	if LEVEL_ERROR >= LEVEL_INFO {
		t.Error("LEVEL_ERROR should be less than LEVEL_INFO")
	}

	if LEVEL_INFO >= LEVEL_DEBUG {
		t.Error("LEVEL_INFO should be less than LEVEL_DEBUG")
	}
}

func TestSetLevel(t *testing.T) {
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	SetLevel(LEVEL_DEBUG)
	if GetLevel() != LEVEL_DEBUG {
		t.Errorf("Expected level %v, got %v", LEVEL_DEBUG, GetLevel())
	}

	SetLevel(LEVEL_SILENT)
	if GetLevel() != LEVEL_SILENT {
		t.Errorf("Expected level %v, got %v", LEVEL_SILENT, GetLevel())
	}
}

func TestIsDebug(t *testing.T) {
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	SetLevel(LEVEL_DEBUG)
	if !IsDebug() {
		t.Error("IsDebug should return true when level is Debug")
	}

	SetLevel(LEVEL_INFO)
	if IsDebug() {
		t.Error("IsDebug should return false when level is Info")
	}
}

func TestIsVerbose(t *testing.T) {
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	SetLevel(LEVEL_DEBUG)
	if IsVerbose() != IsDebug() {
		t.Error("IsVerbose should match IsDebug")
	}

	SetLevel(LEVEL_INFO)
	if IsVerbose() != IsDebug() {
		t.Error("IsVerbose should match IsDebug")
	}
}

func TestEnvironmentVariables(t *testing.T) {
	originalDebug := os.Getenv("GOOGLE_AUTH_WIZARD_DEBUG")
	originalVerbose := os.Getenv("GOOGLE_AUTH_WIZARD_VERBOSE")
	originalSilent := os.Getenv("GOOGLE_AUTH_WIZARD_SILENT")

	defer func() {
		if originalDebug == "" {
			os.Unsetenv("GOOGLE_AUTH_WIZARD_DEBUG")
		} else {
			os.Setenv("GOOGLE_AUTH_WIZARD_DEBUG", originalDebug)
		}

		if originalVerbose == "" {
			os.Unsetenv("GOOGLE_AUTH_WIZARD_VERBOSE")
		} else {
			os.Setenv("GOOGLE_AUTH_WIZARD_VERBOSE", originalVerbose)
		}

		if originalSilent == "" {
			os.Unsetenv("GOOGLE_AUTH_WIZARD_SILENT")
		} else {
			os.Setenv("GOOGLE_AUTH_WIZARD_SILENT", originalSilent)
		}

		globalLogger = &Logger{
			level:  LEVEL_INFO,
			prefix: "[Google Auth Wizard] ",
		}
	}()

	t.Run("Debug Environment", func(t *testing.T) {
		os.Setenv("GOOGLE_AUTH_WIZARD_DEBUG", "true")
		os.Unsetenv("GOOGLE_AUTH_WIZARD_VERBOSE")
		os.Unsetenv("GOOGLE_AUTH_WIZARD_SILENT")

		initializeLogger()

		if GetLevel() != LEVEL_DEBUG {
			t.Errorf("Expected LEVEL_DEBUG, got %v", GetLevel())
		}
	})

	t.Run("Verbose Environment", func(t *testing.T) {
		os.Unsetenv("GOOGLE_AUTH_WIZARD_DEBUG")
		os.Setenv("GOOGLE_AUTH_WIZARD_VERBOSE", "1")
		os.Unsetenv("GOOGLE_AUTH_WIZARD_SILENT")

		initializeLogger()

		if GetLevel() != LEVEL_VERBOSE {
			t.Errorf("Expected LEVEL_VERBOSE for verbose, got %v", GetLevel())
		}
	})

	t.Run("Silent Environment", func(t *testing.T) {
		os.Unsetenv("GOOGLE_AUTH_WIZARD_DEBUG")
		os.Unsetenv("GOOGLE_AUTH_WIZARD_VERBOSE")
		os.Setenv("GOOGLE_AUTH_WIZARD_SILENT", "true")

		initializeLogger()

		if GetLevel() != LEVEL_SILENT {
			t.Errorf("Expected LEVEL_SILENT, got %v", GetLevel())
		}
	})
}

func TestLogFunctions(t *testing.T) {
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	SetLevel(LEVEL_DEBUG)
	Debug("Test debug message")
	Info("Test info message")
	Error("Test error message")
	Print("Test print message")
	Println("Test println message")

	SetLevel(LEVEL_SILENT)
	Debug("This should not appear")
	Info("This should not appear")
	Print("This should not appear")
	Println("This should not appear")
}
