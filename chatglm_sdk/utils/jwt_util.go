package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
	"topic-chain/config"
)

type GLMClaims struct {
	ApiKey    string `json:"api_key"`
	Exp       int64  `json:"exp"`
	Timestamp int64  `json:"timestamp"`
}

// add your own key below
var (
	apiKey = config.LoadStrFromEnv("GLM_API_KEY")
	secret = config.LoadStrFromEnv("GLM_SECRET")
)

func NewGLMToken() string {
	jwtDuration := time.Hour * 24
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"api_key":   apiKey,
		"exp":       time.Now().Add(jwtDuration).UnixNano() / 1e6,
		"timestamp": time.Now().UnixNano() / 1e6,
	})
	token.Header["sign_type"] = "SIGN"
	token.Header["alg"] = "HS256"
	delete(token.Header, "typ")
	secret := []byte(secret)
	signingString, err := token.SignedString(secret)
	if err != nil {
		return ""
	}
	return signingString
}
