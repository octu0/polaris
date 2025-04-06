package polaris

import (
	"fmt"
	"log"
)

type Logger interface {
	DebugF(string, ...any)
	Infof(string, ...any)
	Warnf(string, ...any)
	Errorf(string, ...any)
}

var (
	_ Logger = (*stdLogger)(nil)
)

type stdLogger struct {
	logger    *log.Logger
	debugMode bool
}

// DebugF implements logger.
func (s *stdLogger) DebugF(f string, values ...any) {
	if s.debugMode {
		s.logger.Printf("DEBUG: %s", fmt.Sprintf(f, values...))
	}
}

// Errorf implements logger.
func (s *stdLogger) Errorf(f string, values ...any) {
	s.logger.Printf("ERROR: %s", fmt.Sprintf(f, values...))
}

// Infof implements logger.
func (s *stdLogger) Infof(f string, values ...any) {
	s.logger.Printf("INFO: %s", fmt.Sprintf(f, values...))
}

// Warnf implements logger.
func (s *stdLogger) Warnf(f string, values ...any) {
	s.logger.Printf("WARN: %s", fmt.Sprintf(f, values...))
}
