package config

import (
	"errors"
	"os"
)

type Config struct {
	HTTPPort     string
	DatabaseURL  string
	JWTSecret    string
	YandexAPIKey string
	YandexFolder string
	LogLevel     string
	ReddisAddr   string
}

func Load() (*Config, error) {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		return nil, errors.New("Не указан HTTP_PORT")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("Не указан DATABASE_URL")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("Не указан JWT_SECRET")
	}

	yandexAPIKey := os.Getenv("YANDEX_API_KEY")
	if yandexAPIKey == "" {
		return nil, errors.New("Yandex API Key не указан")
	}

	yandexFolderId := os.Getenv("YANDEX_FOLDER_ID")
	if yandexFolderId == "" {
		return nil, errors.New("Yandex Folder не указан")
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		return nil, errors.New("Loglevel не указан")
	}

	reddisAddr := os.Getenv("REDIS_ADDR")
	if reddisAddr == "" {
		return nil, errors.New("Reddis addr не указан")
	}

	return &Config{
		HTTPPort:     httpPort,
		DatabaseURL:  databaseURL,
		JWTSecret:    jwtSecret,
		YandexAPIKey: yandexAPIKey,
		YandexFolder: yandexFolderId,
		LogLevel:     logLevel,
		ReddisAddr:   reddisAddr,
	}, nil
}
