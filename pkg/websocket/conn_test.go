package websocket

import (
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func Test_newConn(t *testing.T) {
	type args struct {
		ws *websocket.Conn
	}
	tests := []struct {
		name string
		args args
		want net.Conn
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newConn(tt.args.ws); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newConn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wsWrapper_SetDeadline(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		c       *wsWrapper
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.SetDeadline(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("wsWrapper.SetDeadline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_wsWrapper_Write(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		c       *wsWrapper
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("wsWrapper.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("wsWrapper.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wsWrapper_Read(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		c       *wsWrapper
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Read(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("wsWrapper.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("wsWrapper.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wsWrapper_Close(t *testing.T) {
	tests := []struct {
		name    string
		c       *wsWrapper
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Close(); (err != nil) != tt.wantErr {
				t.Errorf("wsWrapper.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
