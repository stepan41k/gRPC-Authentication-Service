package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/stepan41k/gRPC/internal/app/grpc"
	"github.com/stepan41k/gRPC/internal/services/auth"
	"github.com/stepan41k/gRPC/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {

	pool, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, pool, pool, pool, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}