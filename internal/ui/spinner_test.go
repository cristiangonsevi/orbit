package ui

import (
	"testing"
)

func TestColorFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		function func(string) string
		contains string
	}{
		{"ColorRed", "test", ColorRed, "\033[31m"},
		{"ColorGreen", "test", ColorGreen, "\033[32m"},
		{"ColorYellow", "test", ColorYellow, "\033[33m"},
		{"ColorBlue", "test", ColorBlue, "\033[34m"},
		{"ColorMagenta", "test", ColorMagenta, "\033[35m"},
		{"ColorCyan", "test", ColorCyan, "\033[36m"},
		{"ColorBold", "test", ColorBold, "\033[1m"},
		{"ColorDim", "test", ColorDim, "\033[2m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)
			if result == "" {
				t.Errorf("%s returned empty string", tt.name)
			}
			if !contains(result, tt.contains) {
				t.Errorf("%s = %q, want to contain %q", tt.name, result, tt.contains)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSpinnerFrames(t *testing.T) {
	if len(spinnerFrames) == 0 {
		t.Error("spinnerFrames should not be empty")
	}
	if len(spinnerFrames) != 10 {
		t.Errorf("Expected 10 spinner frames, got %d", len(spinnerFrames))
	}
}

func TestNewSpinner(t *testing.T) {
	s := NewSpinner("test message")
	if s == nil {
		t.Error("NewSpinner returned nil")
	}
	if s.message != "test message" {
		t.Errorf("Expected message 'test message', got %q", s.message)
	}
}

func TestProgressBar(t *testing.T) {
	t.Run("new progress bar", func(t *testing.T) {
		pb := NewProgressBar("Test", 10)
		if pb == nil {
			t.Error("NewProgressBar returned nil")
		}
		if pb.total != 10 {
			t.Errorf("Expected total 10, got %d", pb.total)
		}
	})

	t.Run("set total", func(t *testing.T) {
		pb := NewProgressBar("Test", 10)
		pb.SetTotal(20)
		if pb.total != 20 {
			t.Errorf("Expected total 20, got %d", pb.total)
		}
	})

	t.Run("increment", func(t *testing.T) {
		pb := NewProgressBar("Test", 10)
		pb.Increment()
		if pb.finished != 1 {
			t.Errorf("Expected finished 1, got %d", pb.finished)
		}
	})

	t.Run("set progress", func(t *testing.T) {
		pb := NewProgressBar("Test", 10)
		pb.SetProgress(5)
		if pb.finished != 5 {
			t.Errorf("Expected finished 5, got %d", pb.finished)
		}
	})

	t.Run("draw with zero total", func(t *testing.T) {
		pb := NewProgressBar("Test", 0)
		// Should not panic
		pb.Draw()
	})
}

func TestUIConstants(t *testing.T) {
	// Verify color constants are non-empty
	constants := []string{
		colorReset, colorRed, colorGreen, colorYellow,
		colorBlue, colorMagenta, colorCyan, colorWhite,
		colorBold, colorDim,
	}

	for _, c := range constants {
		if c == "" {
			t.Error("Color constant should not be empty")
		}
	}
}
