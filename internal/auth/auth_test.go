package auth

import (
	"context"
	"errors"
	"testing"
)

func TestGetUserID(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ctx     context.Context
		wantID  string
		wantErr error
	}{
		{"present", WithUserID(context.Background(), "alice"), "alice", nil},
		{"missing", context.Background(), "", ErrUserIDNotFound},
		{"wrong type", context.WithValue(context.Background(), userIDKey{}, 42), "", ErrUserIDInvalidType},
		{"empty", WithUserID(context.Background(), ""), "", ErrUserIDEmpty},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id, err := GetUserID(tc.ctx)
			if id != tc.wantID || !errors.Is(err, tc.wantErr) {
				t.Fatalf("GetUserID() = %q, %v; want %q, %v", id, err, tc.wantID, tc.wantErr)
			}
		})
	}
}
