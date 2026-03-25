package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogLevels(t *testing.T) {
	tests := []struct {
		level   Level
		logFunc func(string)
		expect  bool
	}{
		{LevelError, Error, true},
		{LevelError, Warn, false},
		{LevelError, Info, false},
		{LevelError, Debug, false},
		{LevelWarn, Error, true},
		{LevelWarn, Warn, true},
		{LevelWarn, Info, false},
		{LevelWarn, Debug, false},
		{LevelInfo, Error, true},
		{LevelInfo, Warn, true},
		{LevelInfo, Info, true},
		{LevelInfo, Debug, false},
		{LevelDebug, Error, true},
		{LevelDebug, Warn, true},
		{LevelDebug, Info, true},
		{LevelDebug, Debug, true},
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		SetOutput(&buf)
		SetLevel(tt.level)

		tt.logFunc("test message")

		hasOutput := buf.Len() > 0
		if hasOutput != tt.expect {
			t.Errorf("level %v, expected output=%v, got=%v", tt.level, tt.expect, hasOutput)
		}
	}
}

func TestLogFormat(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelInfo)

	Info("test message")

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Error("expected [INFO] in output")
	}
	if !strings.Contains(output, "test message") {
		t.Error("expected message in output")
	}
}

func TestLogWithField(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelInfo)

	WithField("key", "value").Info("test message")

	output := buf.String()
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected key=value in output, got: %s", output)
	}
}

func TestLogWithFields(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelInfo)

	WithFields(map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}).Info("test message")

	output := buf.String()
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("expected key1=value1 in output, got: %s", output)
	}
	if !strings.Contains(output, "key2=42") {
		t.Errorf("expected key2=42 in output, got: %s", output)
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input  string
		expect Level
	}{
		{"error", LevelError},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"info", LevelInfo},
		{"debug", LevelDebug},
		{"unknown", LevelInfo}, // default
	}

	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.expect {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.expect)
		}
	}
}

func TestLogFormatted(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelInfo)

	Infof("hello %s", "world")

	output := buf.String()
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected 'hello world' in output, got: %s", output)
	}
}
