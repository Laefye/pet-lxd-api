package lxd

import (
	"context"
	"fmt"
	"io"

	"github.com/gorilla/websocket"
)

type WebSocketStream struct {
	Conn   *websocket.Conn
	buffer []byte
}

func (r *Rest) wssPath(path resourcePath) string {
	return "wss://" + string(r.Host) + path.String()
}

func (r *Rest) webSocket(ctx context.Context, path resourcePath) (*WebSocketStream, error) {
	wsURL := r.wssPath(path)
	fmt.Println(wsURL)
	conn, _, err := r.Dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, err
	}
	return &WebSocketStream{Conn: conn, buffer: make([]byte, 0)}, nil
}

func (s *WebSocketStream) Close() error {
	return s.Conn.Close()
}

func (s *WebSocketStream) ReadMessage() ([]byte, error) {
	_, message, err := s.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return message, err
}

func (s *WebSocketStream) Read(p []byte) (int, error) {
	if len(s.buffer) == 0 {
		message, err := s.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return 0, err
			}
			return 0, io.EOF
		}
		s.buffer = message
	}
	n := copy(p, s.buffer)
	s.buffer = s.buffer[n:]
	return n, nil
}

func (s *WebSocketStream) Write(message []byte) (int, error) {
	return len(message), s.Conn.WriteMessage(websocket.BinaryMessage, message)
}

type WebSockets map[string]*WebSocketStream

func (ws WebSockets) Close() error {
	for _, stream := range ws {
		if err := stream.Close(); err != nil {
			return err
		}
	}
	return nil
}
