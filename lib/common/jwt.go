package common

import (
	"crypto/rsa"
	"fmt"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var (
	rsaPrivateKeyPath = "private_key.pem"
	rsaPublicKeyPath  = "public_key.pem"
	rsaPrivateKey     *rsa.PrivateKey
	rsaPublicKey      *rsa.PublicKey
)

// Loads private and public keys into memory
func InitJWT() {
	if keyData, err := os.ReadFile(rsaPublicKeyPath); err == nil {
		rsaPublicKey, err = jwt.ParseRSAPublicKeyFromPEM(keyData)
		if err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}
	if keyData, err := os.ReadFile(rsaPrivateKeyPath); err == nil {
		rsaPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyData)
		if err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}
}

// Signs claims and returns the JWT
func SignedJWT(claims jwt.Claims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	s, err := t.SignedString(rsaPrivateKey)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return s, nil
}

// Returns true if JWT could be validated, otherwise returns false
func validateJWT(signedJWT string) jwt.Token {
	token, err := jwt.Parse(signedJWT, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Invalid signing method: %v", token.Header["alg"])
		}
		return rsaPublicKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))

	if err != nil {
		log.Printf("Signature validation error: %s", err.Error())
		return jwt.Token{}
	}
	return *token
}

func ParseJWT(signedJWT string) (claims jwt.MapClaims, valid bool) {
	token := validateJWT(signedJWT)
	if token.Valid != true {
		return nil, false
	}
	claims = token.Claims.(jwt.MapClaims)
	return claims, true
}
