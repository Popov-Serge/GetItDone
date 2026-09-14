package main

import (
	"context"
	"getitdone/internal/config"
	"getitdone/internal/database"
	"getitdone/internal/handler"
	"getitdone/internal/repository"
	"getitdone/internal/router"
	"getitdone/internal/service"
	"getitdone/internal/validation"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
)

func main() {
	configuration, err := config.Load()
	handlerOptions := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}

	logHandler := slog.NewJSONHandler(os.Stdout, &handlerOptions)
	logger := slog.New(logHandler)

	if err != nil {
		logger.Error("Configuration error", "error", err)
		return
	}

	logger.Info("Application started")
	pool, err := database.NewPostgres(configuration.DatabaseURL)
	if err != nil {
		logger.Error("Database connection error", "error", err)
		return
	}
	defer pool.Close()

	redis, err := database.NewRedis(configuration.ReddisAddr)
	if err != nil {
		logger.Error("Reddis connection error", "error", err)
		return
	}
	tokenBlackList := service.NewTokenBlacklist(redis)

	userRepository := repository.NewUserRepository(pool)
	phoneValidator := validation.NewPhoneValidator()
	familyRepository := repository.NewFamilyRepository(pool)
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	requestValidator := handler.NewRequestValidator(validate)

	refreshTokenRepository := repository.NewRefreshTokenRepository(pool)
	jwtService := service.NewJWTService(
		configuration.JWTSecret,
		15*time.Minute,
	)
	refreshTokenService := service.NewRefreshTokenService(
		refreshTokenRepository,
		30*24*time.Hour,
	)

	authService := service.NewAuthService(
		userRepository,
		refreshTokenRepository,
		phoneValidator,
		jwtService,
		refreshTokenService,
		tokenBlackList,
	)

	familyService := service.NewFamilyService(familyRepository)
	authHandler := handler.NewAuthHandler(authService, logger, requestValidator)
	familyHandler := handler.NewFamilyHandler(familyService, logger, requestValidator)
	r := router.New(authHandler, jwtService, tokenBlackList, familyHandler)

	server := &http.Server{
		Addr:              ":" + configuration.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverError := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
		}

		serverError <- err
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverError:
		logger.Error("server stopped", "error", err)

	case sig := <-stop:
		logger.Info("shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("server shutdown failed", "error", err)
		}
	}
}
