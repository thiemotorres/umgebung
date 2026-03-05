package agent

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const timeout = 15 * time.Minute

type request struct {
	Action   string `json:"action"` // "unlock" | "get_key"
	Password string `json:"password,omitempty"`
}

type response struct {
	OK    bool   `json:"ok"`
	Key   []byte `json:"key,omitempty"`
	Error string `json:"error,omitempty"`
}

func SocketPath() string {
	return filepath.Join(fmt.Sprintf("/tmp/umgebung-%d", os.Getuid()), "agent.sock")
}

// RunDaemon starts the agent daemon. Blocks until timeout expires.
func RunDaemon(deriveFn func(password string) ([]byte, error)) error {
	sockPath := SocketPath()
	if err := os.MkdirAll(filepath.Dir(sockPath), 0700); err != nil {
		return err
	}
	os.Remove(sockPath)

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return err
	}
	defer ln.Close()
	defer os.Remove(sockPath)
	os.Chmod(sockPath, 0600)

	var (
		mu    sync.Mutex
		key   []byte
		timer *time.Timer
	)

	resetTimer := func() {
		mu.Lock()
		defer mu.Unlock()
		if timer != nil {
			timer.Reset(timeout)
		}
	}

	timer = time.AfterFunc(timeout, func() {
		ln.Close()
	})

	for {
		conn, err := ln.Accept()
		if err != nil {
			return nil // listener closed = timeout
		}
		go func(c net.Conn) {
			defer c.Close()
			resetTimer()

			var req request
			if err := json.NewDecoder(c).Decode(&req); err != nil {
				return
			}

			var resp response
			switch req.Action {
			case "unlock":
				derived, err := deriveFn(req.Password)
				if err != nil {
					resp = response{OK: false, Error: err.Error()}
				} else {
					mu.Lock()
					key = derived
					mu.Unlock()
					resp = response{OK: true}
				}
			case "get_key":
				mu.Lock()
				k := key
				mu.Unlock()
				if k == nil {
					resp = response{OK: false, Error: "not unlocked"}
				} else {
					resp = response{OK: true, Key: k}
				}
			default:
				resp = response{OK: false, Error: "unknown action"}
			}
			json.NewEncoder(c).Encode(resp)
		}(conn)
	}
}
