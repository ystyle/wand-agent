package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

type Session struct {
	ID     string
	PTY    *os.File
	Cmd    *exec.Cmd
	CWD    string
	mu     sync.Mutex
	closed bool
}

type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{sessions: make(map[string]*Session)}
}

var loginShell string

func init() {
	loginShell = os.Getenv("SHELL")
	if loginShell == "" {
		loginShell = "/bin/bash"
	}
}

func (sm *SessionManager) Create(cwd string, cols, rows int) (*Session, error) {
	cmd := exec.Command(loginShell, "-l")
	if cwd != "" {
		cmd.Dir = cwd
	}
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	fd, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
	if err != nil {
		return nil, fmt.Errorf("pty start: %w", err)
	}

	session := &Session{ID: newID(), PTY: fd, Cmd: cmd, CWD: cwd}

	if cwd == "" {
		link := fmt.Sprintf("/proc/%d/cwd", cmd.Process.Pid)
		if wd, err := os.Readlink(link); err == nil {
			session.CWD = wd
		}
	}

	sm.mu.Lock()
	sm.sessions[session.ID] = session
	sm.mu.Unlock()

	return session, nil
}

func (sm *SessionManager) Get(id string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.sessions[id]
}

func (sm *SessionManager) Remove(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, id)
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.PTY.Close()
}

func (s *Session) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

func newID() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}
