package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",            // Название метрики
			Help: "Общее количество gRPC запрсоов", // Описание метрики
		},
		[]string{"service", "method"}, // Метки для фильтрации запросов по сервису и методу
	)

	// RequestDuration Определяем гистограмму для измерения времени обработки запросов
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds", // Название метрики
			Help:    "Длительность обработки gRPC запросов в секундах",
			Buckets: prometheus.DefBuckets, // Дефолтные интервалы для распределения
		},
		[]string{"service", "method"}, // Метки
	)
)

// InitMetrics Функция инициализации метрик
func InitMetrics(zapLogger *zap.Logger) error {
	// Регистрируем счетчик запросов
	if err := prometheus.Register(RequestCount); err != nil {
		zapLogger.Error("Error to register RequestCount", zap.Error(err))
		return fmt.Errorf("failed to register RequestCount: %w", err)
	}

	// Регистрируем метрику времени запроса
	if err := prometheus.Register(RequestDuration); err != nil {
		zapLogger.Error("error to register RequestDuration", zap.Error(err))
		return fmt.Errorf("failed to register RequestDuration: %w", err)
	}

	return nil
}

// StartPrometheusServer Функция для запуска HTTP-сервера Prometheus
func StartPrometheusServer(zapLogger *zap.Logger, port string) error {
	// Указываем путь для экспорта метрик
	http.Handle("/metrics", promhttp.Handler())

	errorChan := make(chan error)

	go func() {
		// Запускаем HTTP-сервер на порту 9090
		if err := http.ListenAndServe(port, nil); err != nil {
			errorChan <- fmt.Errorf("failed to start HTTP server for Prometheus metrics: %w", err)
		}
	}()

	select {
	case err := <-errorChan:
		zapLogger.Error("Prometheus server error", zap.Error(err))
		return err
	default:
		// Работаем, пока нет ошибок
	}

	return nil
}
