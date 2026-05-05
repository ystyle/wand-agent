package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWS(sm *SessionManager, w http.ResponseWriter, r *http.Request) {
	cols, rows := 80, 24
	cwd := ""
	if err := r.ParseForm(); err == nil {
		if c := r.Form.Get("cols"); c != "" {
			if v, ok := parseInt(c); ok {
				cols = v
			}
		}
		if r := r.Form.Get("rows"); r != "" {
			if v, ok := parseInt(r); ok {
				rows = v
			}
		}
		cwd = r.Form.Get("cwd")
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade: %v", err)
		return
	}
	defer conn.Close()

	session, err := sm.Create(cwd, cols, rows)
	if err != nil {
		log.Printf("create session: %v", err)
		return
	}
	defer cleanupSession(sm, session)

	done := make(chan struct{})

	go ptyToWS(session, conn, done)

	wsToPTY(sm, session, conn, done)
}

func ptyToWS(session *Session, conn *websocket.Conn, done chan struct{}) {
	buf := make([]byte, 4096)
	for {
		n, err := session.PTY.Read(buf)
		if err != nil {
			close(done)
			return
		}
		if session.isClosed() {
			return
		}
		if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
			close(done)
			return
		}
	}
}

func wsToPTY(sm *SessionManager, session *Session, conn *websocket.Conn, done chan struct{}) {
	cols, rows := 80, 24

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if session.isClosed() {
			return
		}

		if isJSONCtrl(msg) {
			if handled := handleCtrl(sm, session, conn, msg, &cols, &rows); handled {
				continue
			}
		}
		session.PTY.Write(msg)
	}
}

func isJSONCtrl(msg []byte) bool {
	return len(msg) > 0 && msg[0] == '{'
}

type ctrlMsg struct {
	Type string `json:"type"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
	CWD  string `json:"cwd,omitempty"`
}

func handleCtrl(sm *SessionManager, session *Session, conn *websocket.Conn, msg []byte, cols, rows *int) bool {
	var ctrl ctrlMsg
	if err := json.Unmarshal(msg, &ctrl); err != nil || ctrl.Type == "" {
		return false
	}

	switch ctrl.Type {
	case "resize":
		if ctrl.Cols > 0 && ctrl.Rows > 0 {
			*cols = ctrl.Cols
			*rows = ctrl.Rows
			pty.Setsize(session.PTY, &pty.Winsize{
				Rows: uint16(ctrl.Rows),
				Cols: uint16(ctrl.Cols),
			})
		}
		return true
	case "ping":
		return true
	case "cwd":
		resp, _ := json.Marshal(map[string]string{"type": "cwd", "dir": session.CWD})
		conn.WriteMessage(websocket.TextMessage, resp)
		return true
	case "fork":
		cwd := session.CWD
		if ctrl.CWD != "" {
			cwd = ctrl.CWD
		}
		forked, err := sm.Create(cwd, *cols, *rows)
		if err != nil {
			resp, _ := json.Marshal(map[string]string{"type": "error", "error": err.Error()})
			conn.WriteMessage(websocket.TextMessage, resp)
			return true
		}
		resp, _ := json.Marshal(map[string]string{"type": "forked", "id": forked.ID})
		conn.WriteMessage(websocket.TextMessage, resp)
		return true
	}
	return false
}

func cleanupSession(sm *SessionManager, session *Session) {
	session.Close()
	session.Cmd.Process.Kill()
	session.Cmd.Wait()
	sm.Remove(session.ID)
}

func parseInt(s string) (int, bool) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}
