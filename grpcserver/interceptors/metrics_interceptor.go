package interceptors

import (
	"context"
	"github.com/arrowwhi/go-utils/grpcserver/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

func MetricsMiddleware(
	serviceName string,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		// Метка для метода
		methodName := info.FullMethod

		// Инкрементируем счетчик запросов
		metrics.RequestCount.WithLabelValues(serviceName, methodName).Inc()

		// Таймер для измерения времени выполнения
		timer := prometheus.NewTimer(
			metrics.RequestDuration.WithLabelValues(serviceName, methodName),
		)
		defer timer.ObserveDuration()

		// Вызываем основной обработчик запроса
		return handler(ctx, req)
	}
}
