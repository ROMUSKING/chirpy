package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	p, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	if err != nil {
		return "", err
	}
	return string(p), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Subject:   userID.String(),
		})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key as a []byte
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("token invalid: %v", err)
	}
	var id uuid.UUID
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		// claims.Subject will give you the subject string
		id, err = uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.Nil, fmt.Errorf("token id invalid: %v", err)
		}
		// Remember you'll need to convert this string back to a UUID
	} else {
		// handle error - claims weren't in the expected format
		return uuid.Nil, fmt.Errorf("token format invalid: %v", err)
	}
	return id, nil

}

func GetBearerToken(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("no auth token in the header")
	}
	token = strings.TrimPrefix(token, "Bearer")
	token = strings.Trim(token, " 	")
	return token, nil
}

func MakeRefreshToken() (string, error) {

	c := 32
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	return token, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("no auth token in the header")
	}
	token = strings.TrimPrefix(token, "ApiKey")
	token = strings.Trim(token, " 	")
	return token, nil
}
