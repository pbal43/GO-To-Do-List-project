package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type HS256Signer struct {
	Secret     []byte
	Issuer     string
	Audience   string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func generateJti() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (hs HS256Signer) NewAccessToken(userID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    hs.Issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{hs.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(hs.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateJti(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["typ"] = "JWT"
	return token.SignedString(hs.Secret)
}

func (hs HS256Signer) NewRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    hs.Issuer,
		Subject:   userID,
		Audience:  jwt.ClaimStrings{hs.Audience},
		ExpiresAt: jwt.NewNumericDate(now.Add(hs.RefreshTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        generateJti(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["typ"] = "JWT"
	return token.SignedString(hs.Secret)
}

func (hs HS256Signer) GetIssuer() string {
	return hs.Issuer
}

func (hs HS256Signer) GetAudience() string {
	return hs.Audience
}
