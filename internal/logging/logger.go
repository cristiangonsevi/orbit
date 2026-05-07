package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	logDir      = ".local/share/orbit/logs"
	maxLogFiles = 30
)

// Entry represents a single log entry for one workflow execution.
type Entry struct {
	Timestamp   string   `json:"timestamp"`
	ProjectName string   `json:"project_name"`
	UserHost    string   `json:"user_host"`
	Commands    []string `json:"commands,omitempty"`
	Success     bool     `json:"success"`
	Duration    string   `json:"duration"`
	Error       string   `json:"error,omitempty"`
}

// Logger handles persistent execution logs.
type Logger struct {
	logFile *os.File
}

// New creates or opens the log file for this execution.
// It also performs log rotation by removing old files beyond maxLogFiles.
func New(projectName, userHost string, commands []string) (*Logger, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	logDirPath := filepath.Join(home, logDir)
	if err := os.MkdirAll(logDirPath, 0755); err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// Rotate old logs
	if err := rotateLogs(logDirPath); err != nil {
		// Non-fatal: log the error but continue
		fmt.Fprintf(os.Stderr, "[LOG] Warning: log rotation failed: %v\n", err)
	}

	timestamp := time.Now().Format("2006-01-02T15-04-05")
	logFileName := fmt.Sprintf("%s_%s.json", projectName, timestamp)
	logFilePath := filepath.Join(logDirPath, logFileName)

	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}

	return &Logger{logFile: f}, nil
}

// Log writes an entry to the log file.
func (l *Logger) Log(entry Entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling log entry: %w", err)
	}
	if _, err := l.logFile.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("writing log entry: %w", err)
	}
	return nil
}

// Close closes the log file.
func (l *Logger) Close() error {
	return l.logFile.Close()
}

// rotateLogs removes log files beyond maxLogFiles, keeping the most recent.
func rotateLogs(logDirPath string) error {
	entries, err := os.ReadDir(logDirPath)
	if err != nil {
		return err
	}

	var logFiles []os.FileInfo
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			info, err := e.Info()
			if err != nil {
				continue
			}
			logFiles = append(logFiles, info)
		}
	}

	if len(logFiles) <= maxLogFiles {
		return nil
	}

	// Sort by modification time, oldest first
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].ModTime().Before(logFiles[j].ModTime())
	})

	toDelete := len(logFiles) - maxLogFiles
	for i := 0; i < toDelete; i++ {
		os.Remove(filepath.Join(logDirPath, logFiles[i].Name()))
	}

	return nil
}
