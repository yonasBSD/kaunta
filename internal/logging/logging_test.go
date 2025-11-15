package logging

import (
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetLoggerForTest() {
	initOnce = sync.Once{}
	logger = nil
	exitFunc = os.Exit
}

func TestParseLevelMappings(t *testing.T) {
	assert.Equal(t, slog.LevelDebug, parseLevel("debug"))
	assert.Equal(t, slog.LevelWarn, parseLevel("warn"))
	assert.Equal(t, slog.LevelWarn, parseLevel("warning"))
	assert.Equal(t, slog.LevelError, parseLevel("error"))
	assert.Equal(t, slog.LevelInfo, parseLevel("unknown"))
}

func TestLoggerSingleton(t *testing.T) {
	resetLoggerForTest()
	first := L()
	second := L()
	assert.Same(t, first, second)
}

func TestFatalInvokesExitFunction(t *testing.T) {
	resetLoggerForTest()

	var exitCode int
	exitFunc = func(code int) {
		exitCode = code
	}

	// Replace logger with one writing to /dev/null to avoid noisy output
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	initOnce = sync.Once{} // prevent L() from reinitializing
	initOnce.Do(func() {}) // mark as done so L() uses existing logger

	Fatal("boom", "key", "value")

	require.Equal(t, 1, exitCode)
}
