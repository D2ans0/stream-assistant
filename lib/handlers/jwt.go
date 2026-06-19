package handlers

import (
	"crypto/rsa"
	"fmt"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var hmacSampleSecret []byte
var rsaPrivateKey *rsa.PrivateKey
var rsaPublicKey *rsa.PublicKey

func Init() {
	if keyData, err := os.ReadFile("public_key.pem"); err == nil {
		rsaPublicKey, err = jwt.ParseRSAPublicKeyFromPEM(keyData)
		if err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}
	if keyData, err := os.ReadFile("private_key.pem"); err == nil {
		rsaPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyData)
		if err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}
}

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
func ValidateJWT(signedJWT string) bool {
	token, err := jwt.Parse(signedJWT, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Invalid signing method: %v", token.Header["alg"])
		}
		return rsaPublicKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))

	if err != nil {
		log.Printf("Signature validation error: %s", err.Error())
		return false
	}

	if token.Valid {
		return true
	}
	return false
}
