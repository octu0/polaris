package polaris

import (
	"bytes"
	"log"
	"testing"
)

func newTestStdLogger(debugMode bool) (*stdLogger, *bytes.Buffer) {
	buf := bytes.NewBuffer(nil)
	testLogger := log.New(buf, "", 0)
	logger := &stdLogger{
		logger:    testLogger,
		debugMode: debugMode,
	}
	return logger, buf
}

func TestStdLoggerDebugMode(t *testing.T) {
	t.Run("enable", func(tt *testing.T) {
		logger, buf := newTestStdLogger(true)
		format := "Test debug message %d with string %s"
		args := []any{123, "hello"}
		expectedPrefix := "DEBUG: "
		expectedMsg := "Test debug message 123 with string hello"

		logger.Debugf(format, args...)

		actualOutput := buf.String()
		expectedOutput := expectedPrefix + expectedMsg + "\n" // new line add logger

		if actualOutput != expectedOutput {
			tt.Errorf("Debugf (enabled) output mismatch:\nExpected: %q\nActual:   %q", expectedOutput, actualOutput)
		}
	})
	t.Run("disable", func(tt *testing.T) {
		logger, buf := newTestStdLogger(false)
		format := "This should not be logged %s"
		args := []any{"ever"}

		logger.Debugf(format, args...)

		actualOutput := buf.String()
		expectedOutput := ""

		if actualOutput != expectedOutput {
			tt.Errorf("Debugf (disabled) should produce no output, but got: %q", actualOutput)
		}
	})
}

func TestStdLogger_Infof(t *testing.T) {
	logger, buf := newTestStdLogger(false) // debugMode doesn't affect Info
	format := "Info message: %v"
	args := []any{map[string]int{"a": 1}}
	expectedPrefix := "INFO: "
	expectedMsg := "Info message: map[a:1]"

	logger.Infof(format, args...)

	actualOutput := buf.String()
	expectedOutput := expectedPrefix + expectedMsg + "\n"

	if actualOutput != expectedOutput {
		t.Errorf("Infof output mismatch:\nExpected: %q\nActual:   %q", expectedOutput, actualOutput)
	}
}

func TestStdLogger_Warnf(t *testing.T) {
	logger, buf := newTestStdLogger(false) // debugMode doesn't affect Warn
	format := "Warning! Value is %f"
	args := []any{3.14}
	expectedPrefix := "WARN: "
	expectedMsg := "Warning! Value is 3.140000"

	logger.Warnf(format, args...)

	actualOutput := buf.String()
	expectedOutput := expectedPrefix + expectedMsg + "\n"

	if actualOutput != expectedOutput {
		t.Errorf("Warnf output mismatch:\nExpected: %q\nActual:   %q", expectedOutput, actualOutput)
	}
}

func TestStdLogger_Errorf(t *testing.T) {
	logger, buf := newTestStdLogger(false) // debugMode doesn't affect Error
	format := "Error occurred: %s - Code: %d"
	args := []any{"File not found", 404}
	expectedPrefix := "ERROR: "
	expectedMsg := "Error occurred: File not found - Code: 404"

	logger.Errorf(format, args...)

	actualOutput := buf.String()
	expectedOutput := expectedPrefix + expectedMsg + "\n"

	if actualOutput != expectedOutput {
		t.Errorf("Errorf output mismatch:\nExpected: %q\nActual:   %q", expectedOutput, actualOutput)
	}
}

func TestStdLogger_NoArgs(t *testing.T) {
	logger, buf := newTestStdLogger(false)
	expectedPrefix := "INFO: "
	expectedMsg := "Simple message"

	logger.Infof("Simple message") // Call without args...

	actualOutput := buf.String()
	expectedOutput := expectedPrefix + expectedMsg + "\n"

	if actualOutput != expectedOutput {
		t.Errorf("Infof (no args) output mismatch:\nExpected: %q\nActual:   %q", expectedOutput, actualOutput)
	}
}
