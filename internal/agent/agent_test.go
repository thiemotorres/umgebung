package agent_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/thiemotorres/umgebung/internal/agent"
)

func TestDaemonUnlockAndGetKey(t *testing.T) {
	// Use a unique socket path per test to avoid conflicts
	t.Setenv("HOME", t.TempDir()) // forces a different uid-based path won't work, use override

	expectedKey := []byte("derived-key-32-bytes-exactly!!!!")
	deriveFn := func(pw string) ([]byte, error) {
		if pw == "correct" {
			return expectedKey, nil
		}
		return nil, nil
	}

	// Clean up any leftover socket
	os.Remove(agent.SocketPath())

	done := make(chan error, 1)
	go func() {
		done <- agent.RunDaemon(deriveFn)
	}()

	// Wait for socket to appear
	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)
		if _, err := os.Stat(agent.SocketPath()); err == nil {
			break
		}
	}

	// Test unlock
	if err := agent.Unlock("correct"); err != nil {
		t.Fatalf("Unlock: %v", err)
	}

	// Test get key
	key, err := agent.GetKey()
	if err != nil {
		t.Fatalf("GetKey: %v", err)
	}
	if !bytes.Equal(key, expectedKey) {
		t.Fatalf("got %v, want %v", key, expectedKey)
	}

	// Clean up: remove socket to stop daemon
	os.Remove(agent.SocketPath())
}
