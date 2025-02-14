package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/stepan41k/gRPC/internal/app"
	"github.com/stepan41k/gRPC/internal/config"
	"github.com/stepan41k/gRPC/internal/lib/logger/slogpretty"
)

const (
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application", slog.Any("config", cfg))

	storage := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", cfg.Storage.Host, cfg.Storage.Port, cfg.Storage.Username, cfg.Storage.DBName, os.Getenv("DB_PASSWORD"), cfg.Storage.SSLMode)

	application := app.New(log, cfg.GRPC.Port, storage, cfg.TokenTTL)


	go func() {
		application.GRPCServer.MustRun()
	}()
	
	//Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	signal := <-stop

	log.Info("stopping application", slog.String("signal", signal.String()))

	application.GRPCServer.Stop()

	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

		switch env {
		case envLocal:
			log = setupPrettySlog()
		case envDev:
			log = slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
		case envProd:
			log = slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
			)
		}
	
	return log
}

func setupPrettySlog() *slog.Logger {
	opts :=	slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}