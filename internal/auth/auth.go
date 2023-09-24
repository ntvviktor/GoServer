package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
)

const (
	AccessToken  = "access"
	RefreshToken = "refresh"
)

func HashUserPassword(password string) (string, error) {
	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(pw), err
}

func AuthenticateUser(login string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(password), []byte(login))
}

func GenerateJWT(id int, jwtSecret string, tokenType string) (string, error) {
	var issuer string
	var expireIn time.Duration
	switch tokenType {
	case "access":
		issuer = "chirpy-access"
		expireIn = time.Hour * 1
	case "refresh":
		issuer = "chirpy-refresh"
		expireIn = time.Hour * 24 * 60
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expireIn)),
		Subject:   fmt.Sprintf("%d", id),
	})
	return token.SignedString([]byte(jwtSecret))
}

func ValidateJWT(tokenString string, jwtSecret string) (int, string, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(jwtSecret), nil },
	)
	if err != nil {
		return -1, "", err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return -1, "", err
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return -1, "", err
	}

	validID, err := strconv.Atoi(id)
	if err != nil {
		return -1, "", err
	}

	return validID, issuer, nil
}
