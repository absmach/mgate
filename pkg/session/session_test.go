package session

import (
	"context"
	"reflect"
	"testing"
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewContext(tt.args.ctx, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewContext() = %v, want %v", got, tt.want)
			}
		})
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := FromContext(tt.args.ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromContext() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FromContext() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
