package websocket

import (
	"fmt"
	"net/http"
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
	wsConn, _, err := websocket.DefaultDialer.Dial(testURL, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer wsConn.Close()

	tests := []struct {
		name     string
		client   *wsWrapper
		deadline time.Time
		wantErr  bool
	}{
		{
			name:     "Successfully set deadline",
			client:   &wsWrapper{Conn: wsConn},
			deadline: time.Now(),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		err := tt.client.SetDeadline(tt.deadline)
		assert.Equal(t, err != nil, tt.wantErr, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.wantErr, err))
	}
}

func Test_Write(t *testing.T) {
	wsConn, _, err := websocket.DefaultDialer.Dial(testURL, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer wsConn.Close()

	tests := []struct {
		name   string
		client *wsWrapper
		data   []byte
		resp   int
		err    error
	}{
		{
			name: "Successfully write data",
			client: &wsWrapper{
				Conn: wsConn,
			},
			data: []byte("test"),
			resp: 4,
			err:  nil,
		},
	}

	for _, tt := range tests {
		got, err := tt.client.Write(tt.data)
		assert.Equal(t, got, tt.resp, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.resp, got))
		assert.Equal(t, err, tt.err, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.err, err))
	}
}

func Test_Read(t *testing.T) {
	wsConn, _, err := websocket.DefaultDialer.Dial(testURL, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer wsConn.Close()

	go func() {
		for {
			client := &wsWrapper{
				Conn: wsConn,
			}
			_, err := client.Write([]byte("test"))
			if err != nil {
				return
			}

			time.Sleep(1 * time.Second)
		}
	}()

	tests := []struct {
		name   string
		client *wsWrapper
		data   []byte
		resp   int
		err    bool
	}{
		{
			name: "Successfully read data",
			client: &wsWrapper{
				Conn: wsConn,
			},
			data: []byte("test"),
			resp: 4,
			err:  false,
		},
	}
	for _, tt := range tests {
		got, err := tt.client.Read(tt.data)
		assert.Equal(t, got, tt.resp, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.resp, got))
		require.Nil(t, err, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.err, err))
	}
}

func Test_Close(t *testing.T) {
	wsConn, _, err := websocket.DefaultDialer.Dial(testURL, nil)
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

func TestNewConn(t *testing.T) {
	wsConn, _, err := websocket.DefaultDialer.Dial(testURL, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer wsConn.Close()

	tests := []struct {
		name string
		conn *websocket.Conn
	}{
		{
			name: "Successfully created new connection",
			conn: wsConn,
		},
	}

	for _, tt := range tests {
		got := newConn(tt.conn)
		assert.NotNil(t, got, fmt.Sprintf("%s: expected %v got %v\n", tt.name, got, nil))
	}
}
