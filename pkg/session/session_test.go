package session

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	type args struct {
		ctx context.Context
		s   *Session
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "successfully created new context",
			args: args{
				ctx: context.Background(),
				s: &Session{
					ID:       "myID",
					Username: "myName",
					Password: nil,
				},
			},
			want: context.WithValue(context.Background(), sessionKey{}, &Session{
				ID:       "myID",
				Username: "myName",
				Password: nil,
			}),
		},
	}

	for _, tt := range tests {
		got := NewContext(tt.args.ctx, tt.args.s)
		assert.Equal(t, got, tt.want, fmt.Sprintf("%s: expected %s got %s\n", tt.name, tt.want, got))
	}
}

func TestFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name  string
		args  args
		want  *Session
		want1 bool
	}{
		{
			name: "successfully get session from context",
			args: args{
				context.WithValue(context.TODO(), sessionKey{}, &Session{
					ID:       "myID",
					Username: "myName",
					Password: nil,
				}),
			},
			want: &Session{
				ID:       "myID",
				Username: "myName",
				Password: nil,
			},
			want1: true,
		},
	}

	for _, tt := range tests {
		got, gotBool := FromContext(tt.args.ctx)
		assert.Equal(t, got.ID, tt.want.ID, fmt.Sprintf("%s: expected %s got %s\n", tt.name, tt.want.ID, got.ID))
		assert.Equal(t, got.Password, tt.want.Password, fmt.Sprintf("%s: expected %s got %s\n", tt.name, tt.want.Password, got.Password))
		assert.True(t, gotBool == tt.want1, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.want1, gotBool))
	}
}
