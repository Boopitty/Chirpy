package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Change a password into a hash before it's stored in the database
func HashPassword(password string) (string, error) {
	param := argon2id.DefaultParams
	hash, err := argon2id.CreateHash(password, param)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// Check the validity of a hashed password
func CheckPasswordHash(password, hash string) (bool, error) {
	check, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return check, nil
}

// Makes a JSON Web Token for a user
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	if userID == uuid.Nil {
		return "", errors.New("invalid user id")
	}

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := jwtToken.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

// Validates a JWT and returns the user id from the subject claim if the token is valid.
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	// Claims struct for storing info parsed with jwt.ParseWithClaims
	claims := &jwt.RegisteredClaims{}

	// Returns the same key type ([]byte) used when the token was signed.
	// An error will be returned if the token is invalid or has expired.
	keyFunc := func(token *jwt.Token) (any, error) {
		// Ensure signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		return []byte(tokenSecret), nil
	}

	// Validate the signature of the JWT
	// and extract the claims into a *jwt.Token struct
	_, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return uuid.Nil, err
	}

	// Extract user id from subject claim
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// Extract the bearer token from the authorization header of a request
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}
	stripedHeader := strings.TrimPrefix(authHeader, "Bearer ")
	return stripedHeader, nil
}

// Make a random refresh token for a user
func MakeRefreshToken() string {
	t := make([]byte, 32)
	rand.Read(t)
	token := hex.EncodeToString(t)
	return token
}

func GetAPIKey(headers http.Header) (string, error) {
	apiKey := headers.Get("Authorization")
	if apiKey == "" {
		return "", errors.New("missing API key header")
	}
	return strings.TrimPrefix(apiKey, "ApiKey "), nil
}
