package agent

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// EnsureRunning starts the daemon if not already running.
// daemonBin is the path to the current executable (os.Executable()).
func EnsureRunning(daemonBin string) error {
	if isRunning() {
		return nil
	}
	cmd := exec.Command(daemonBin, "__agent")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start agent: %w", err)
	}
	// Wait for socket to appear
	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)
		if isRunning() {
			return nil
		}
	}
	return fmt.Errorf("agent did not start")
}

func isRunning() bool {
	_, err := os.Stat(SocketPath())
	return err == nil
}

func send(req request) (response, error) {
	conn, err := net.DialTimeout("unix", SocketPath(), 2*time.Second)
	if err != nil {
		return response{}, fmt.Errorf("connect to agent: %w", err)
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return response{}, err
	}
	var resp response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return response{}, err
	}
	return resp, nil
}

// Unlock sends the password to the daemon for key derivation.
func Unlock(password string) error {
	resp, err := send(request{Action: "unlock", Password: password})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("unlock failed: %s", resp.Error)
	}
	return nil
}

// GetKey retrieves the derived key from the daemon.
func GetKey() ([]byte, error) {
	resp, err := send(request{Action: "get_key"})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("get key: %s", resp.Error)
	}
	return resp.Key, nil
}
