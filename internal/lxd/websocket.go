package lxd

import (
	"context"
	"io"

	"github.com/gorilla/websocket"
)

type WebSocketStream struct {
	Conn   *websocket.Conn
	buffer []byte
}

func ConnectWebsocket(ctx context.Context, dialer websocket.Dialer, endpoint Endpoint, path Path) (*WebSocketStream, error) {
	wsURL := endpoint.Wss(path.String())
	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, err
	}
	stream := &WebSocketStream{
		Conn:   conn,
		buffer: make([]byte, 0),
	}
	return stream, nil
}

func (s *WebSocketStream) Close() error {
	return s.Conn.Close()
}

func (s *WebSocketStream) ReadMessage() ([]byte, error) {
	_, message, err := s.Conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			return nil, err
		}
		return nil, io.EOF
	}
	return message, err
}

func (s *WebSocketStream) Read(p []byte) (int, error) {
	if len(s.buffer) == 0 {
		message, err := s.ReadMessage()
		if err != nil {
			return 0, err
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
