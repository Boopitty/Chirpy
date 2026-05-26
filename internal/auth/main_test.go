package auth_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Boopitty/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func TestHashAndCheckPassword(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		password      string
		checkPassword string
		want          bool
		wantHashErr   bool
		wantCheckErr  bool
	}{
		{
			name:          "Valid Password",
			password:      "password",
			checkPassword: "password",
			want:          true,
			wantHashErr:   false,
			wantCheckErr:  false,
		},
		{
			name:          "Wrong Password",
			password:      "password",
			checkPassword: "wrong-pasword",
			want:          false,
			wantHashErr:   false,
			wantCheckErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// hash the password
			hash, hashErr := auth.HashPassword(tt.password)
			if hashErr != nil {
				if !tt.wantHashErr {
					t.Errorf("failed to hash password: %v", hashErr)
				}
				return
			}
			if tt.wantHashErr {
				t.Fatalf("HashPassword() succeeded unexpectedly")
			}

			if hash == tt.password {
				t.Errorf("HashPassword() = %v, want other value", hash)
			}

			// Check the password
			got, gotErr := auth.CheckPasswordHash(tt.checkPassword, hash)
			if gotErr != nil {
				if !tt.wantCheckErr {
					t.Errorf("CheckPasswordHash() failed: %v", gotErr)
				}
				return
			}
			if tt.wantCheckErr {
				t.Fatal("CheckPasswordHash() succeeded unexpectedly")
			}

			if got != tt.want {
				t.Errorf("CheckPasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeJWTAndValidate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		userID          uuid.UUID
		makeSecret      string
		validateSecret  string
		expiresIn       time.Duration
		wantMakeErr     bool
		wantValidateErr bool
		wantIdErr       bool
	}{
		{
			name:            "Valid Token",
			userID:          uuid.New(),
			makeSecret:      "super-secret-key",
			validateSecret:  "super-secret-key",
			expiresIn:       time.Second,
			wantMakeErr:     false,
			wantValidateErr: false,
			wantIdErr:       false,
		},
		{
			name:            "Time Expired",
			userID:          uuid.New(),
			makeSecret:      "secret",
			validateSecret:  "secret",
			expiresIn:       -time.Second,
			wantMakeErr:     false,
			wantValidateErr: true,
			wantIdErr:       false,
		},
		{
			name:            "Wrong Secret",
			userID:          uuid.New(),
			makeSecret:      "secret",
			validateSecret:  "wrong-secret",
			expiresIn:       time.Second,
			wantMakeErr:     false,
			wantValidateErr: true,
			wantIdErr:       false,
		},
		{
			name:            "Empty User ID",
			userID:          uuid.Nil,
			makeSecret:      "secret",
			validateSecret:  "secret",
			expiresIn:       time.Second,
			wantMakeErr:     true,
			wantValidateErr: false,
			wantIdErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := auth.MakeJWT(tt.userID, tt.makeSecret, tt.expiresIn)
			if gotErr != nil {
				if !tt.wantMakeErr {
					t.Errorf("MakeJWT() failed: %v", gotErr)
				}
				return
			}
			if tt.wantMakeErr {
				t.Fatal("MakeJWT() succeeded unexpectedly")
			}

			gotID, err := auth.ValidateJWT(got, tt.validateSecret)
			if err != nil {
				if !tt.wantValidateErr {
					t.Errorf("ValidateJWT() failed: %v", err)
				}
				return
			}
			if tt.wantValidateErr {
				t.Fatal("ValidateJWT() succeeded unexpectedly")
			}
			if gotID != tt.userID {
				if !tt.wantIdErr {
					t.Errorf("ValidateJWT() = %v, want: %v", got, tt.userID)
				}
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		headers http.Header
		want    string
		wantErr bool
	}{
		{
			name: "Valid Bearer Token",
			headers: http.Header{
				"Authorization": []string{"Bearer valid-token"},
			},
			want:    "valid-token",
			wantErr: false,
		},
		{
			name: "Missing Authorization Header",
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := auth.GetBearerToken(tt.headers)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetBearerToken() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetBearerToken() succeeded unexpectedly")
			}

			if got != tt.want {
				t.Errorf("GetBearerToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
