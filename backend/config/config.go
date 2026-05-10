package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	KeycloakBase  string
	ClientID      string
	ClientSecret  string
	RedirectURI   string
	PostLogoutURI string
	FrontendURL   string
	RedisURL      string
}

var Cfg Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("找不到 .env 檔案")
	}

	Cfg = Config{
		KeycloakBase:  os.Getenv("KEYCLOAK_BASE"),
		ClientID:      os.Getenv("CLIENT_ID"),
		ClientSecret:  os.Getenv("CLIENT_SECRET"),
		RedirectURI:   os.Getenv("REDIRECT_URI"),
		PostLogoutURI: os.Getenv("POST_LOGOUT_URI"),
		FrontendURL:   os.Getenv("FRONTEND_URL"),
		RedisURL:      os.Getenv("REDIS_URL"),
	}
}
