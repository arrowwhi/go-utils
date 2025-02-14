package main

import (
	"context"
	"fmt"
	"github.com/arrowwhi/go-utils/grpcserver"
	"github.com/arrowwhi/go-utils/grpcserver/test/config"
	"github.com/arrowwhi/go-utils/grpcserver/test/handler"
	"github.com/arrowwhi/go-utils/logger"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// Загрузка конфигурации
	cfg, err := config.GetConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %s\n", err.Error())
	}

	// Инициализация логгера
	zapLogger := logger.NewClientZapLogger(cfg.LogLevel, cfg.ServerConfig.ServiceName)

	srv, err := grpcserver.NewServer(
		cfg.ServerConfig,
		zapLogger,
		grpcserver.WithImplementationAdapters(
			handler.New(zapLogger),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create server: %s\n", err.Error())
	}

	// Обработка системных сигналов для корректного завершения работы
	quit := setupSignalChannel()
	serverErrors := make(chan error, 1)

	zapLogger.Info("Starting gRPC server", zap.String("gRPC port", cfg.ServerConfig.GRPCPort))

	// Запуск gRPC сервера в горутине
	go func() {
		serverErrors <- srv.Start(context.Background())
	}()

	// Ожидание ошибки сервера или сигнала завершения
	select {
	case err = <-serverErrors:
		panic(err)
	case sig := <-quit:
		zapLogger.Info(fmt.Sprintf("Received termination signal: %s", sig))
	}

	// Корректное завершение работы gRPC сервера
	zapLogger.Info("Shutting down gRPC server gracefully...")
	srv.Stop()
	zapLogger.Info("gRPC server stopped")
}

func setupSignalChannel() chan os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	return quit
}
