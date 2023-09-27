package websocket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	msgChan = make(chan []byte)
	count   uint64
)

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		atomic.AddUint64(&count, 1)
		msgChan <- message
	}
}

func Test_SetDeadline(t *testing.T) {
	type args struct {
		t time.Time
	}

	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	u := strings.Replace(s.URL, "http", "ws", 1)

	// Connect to the server
	wsConn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer wsConn.Close()

	tests := []struct {
		name    string
		c       *wsWrapper
		args    args
		wantErr error
	}{
		{
			name: "Successfully set deadline",
			c: &wsWrapper{
				Conn: wsConn,
			},
			args: args{
				t: time.Now(),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		err := tt.c.SetDeadline(tt.args.t)
		assert.Equal(t, err, tt.wantErr, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.wantErr, err))
	}
}

func Test_Write(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	u := strings.Replace(s.URL, "http", "ws", 1)

	// Connect to the server
	wsConn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer wsConn.Close()

	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		c       *wsWrapper
		args    args
		want    int
		wantErr error
	}{
		{
			name: "Successfully wrote data",
			c: &wsWrapper{
				Conn: wsConn,
			},
			args: args{
				p: []byte("test"),
			},
			want:    4,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		got, err := tt.c.Write(tt.args.p)
		assert.Equal(t, got, tt.want, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.want, got))
		assert.Equal(t, err, tt.wantErr, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.wantErr, err))
	}
}

func Test_Read(t *testing.T) {
	type args struct {
		p []byte
	}

	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	u := "wss://echo.websocket.org"

	// Connect to the server
	wsConn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer wsConn.Close()

	tests := []struct {
		name    string
		c       *wsWrapper
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Successfully read data",
			c: &wsWrapper{
				Conn: wsConn,
			},
			args: args{
				p: []byte("test"),
			},
			want:    4,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := tt.c.Read(tt.args.p)
		assert.Equal(t, got, tt.want, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.want, got))
		require.Nil(t, err, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.wantErr, err))
	}
}

func Test_Close(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	u := strings.Replace(s.URL, "http", "ws", 1)

	// Connect to the server
	wsConn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer wsConn.Close()

	tests := []struct {
		name    string
		c       *wsWrapper
		wantErr error
	}{
		{
			name: "Successfully closed connection",
			c: &wsWrapper{
				Conn: wsConn,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		err := tt.c.Close()
		require.Nil(t, err, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.wantErr, err))
	}
}
