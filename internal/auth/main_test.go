package auth_test

import (
	"testing"
	"time"

	"github.com/Boopitty/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		password string
		wantErr  bool
	}{
		{
			name:     "Normal case",
			password: "password",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := auth.HashPassword(tt.password)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("HashPassword() failed: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("HashPassword() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if tt.password == got {
				t.Errorf("HashPassword() = %v, password has not changed", got)
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		password string
		hash     string
		want     bool
		wantErr  bool
	}{
		{
			name:     "Normal case",
			password: "coolpassword",
			want:     true,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// hash the password
			var hashErr error
			tt.hash, hashErr = auth.HashPassword(tt.password)
			if hashErr != nil {
				t.Errorf("failed to hash password: %v", hashErr)
				return
			}

			got, gotErr := auth.CheckPasswordHash(tt.password, tt.hash)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CheckPasswordHash() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CheckPasswordHash() succeeded unexpectedly")
			}

			if got != tt.want {
				t.Errorf("CheckPasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeJWT(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		userID      uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
		want        string
		wantErr     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := auth.MakeJWT(tt.userID, tt.tokenSecret, tt.expiresIn)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("MakeJWT() failed: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("MakeJWT() succeeded unexpectedly")
			}

			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("MakeJWT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		tokenString string
		tokenSecret string
		want        uuid.UUID
		wantErr     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := auth.ValidateJWT(tt.tokenString, tt.tokenSecret)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateJWT() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateJWT() succeeded unexpectedly")
			}

			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("ValidateJWT() = %v, want %v", got, tt.want)
			}
		})
	}
}
