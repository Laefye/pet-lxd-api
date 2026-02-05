package lxd

import (
	"context"
	"io"

	"github.com/gorilla/websocket"
)

type WebSocketStream struct {
	Conn *websocket.Conn
}

func ConnectWebsocket(ctx context.Context, dialer websocket.Dialer, endpoint Endpoint, path Path) (*WebSocketStream, error) {
	wsURL := endpoint.Wss(path.String())
	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, err
	}
	stream := &WebSocketStream{
		Conn: conn,
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

func (s *WebSocketStream) Write(message []byte) (int, error) {
	return len(message), s.Conn.WriteMessage(websocket.BinaryMessage, message)
}
