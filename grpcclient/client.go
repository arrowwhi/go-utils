package grpcclient

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Client представляет gRPC клиент с настраиваемыми опциями
type Client struct {
	conn           *grpc.ClientConn
	timeout        time.Duration
	maxRetries     int
	methodHandlers map[string]MethodHandler
	metadata       metadata.MD
}

// ClientOption определяет функцию для настройки клиента
type ClientOption func(*Client)

// MethodHandler определяет обработчик метода gRPC
type MethodHandler func(ctx context.Context, req interface{}) (interface{}, error)

// NewClient создает новый gRPC клиент
func NewClient(target string, opts ...ClientOption) (*Client, error) {
	c := &Client{
		timeout:        5 * time.Second,
		maxRetries:     1,
		methodHandlers: make(map[string]MethodHandler),
		metadata:       metadata.MD{},
	}

	// Применяем опции конфигурации
	for _, opt := range opts {
		opt(c)
	}

	// Создаем соединение с дополнительными опциями
	var err error
	c.conn, err = grpc.Dial(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания соединения: %w", err)
	}

	return c, nil
}

// WithTimeout устанавливает таймаут для клиента
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithMaxRetries устанавливает максимальное количество повторных попыток
func WithMaxRetries(maxRetries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// AddHandler добавляет пользовательский обработчик для метода
func (c *Client) AddHandler(method string, handler MethodHandler) {
	c.methodHandlers[method] = handler
}

// Request представляет gRPC запрос
type Request struct {
	client   *Client
	method   string
	body     interface{}
	response interface{}
	ctx      context.Context
	metadata metadata.MD
}

// NewRequest создает новый запрос
func NewRequest(client *Client, method string, body interface{}, response interface{}) *Request {
	return &Request{
		client:   client,
		method:   method,
		body:     body,
		response: response,
		metadata: client.metadata,
	}
}

// WithMetadata добавляет метаданные к запросу
func (r *Request) WithMetadata(md metadata.MD) *Request {
	r.metadata = md
	return r
}

// WithContext устанавливает контекст для запроса
func (r *Request) WithContext(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

// Do выполняет gRPC запрос с поддержкой повторных попыток
func (r *Request) Do() (interface{}, error) {
	if r.ctx == nil {
		r.ctx = context.Background()
	}

	if len(r.metadata) > 0 {
		r.ctx = metadata.NewOutgoingContext(r.ctx, r.metadata)
	}

	ctx, cancel := context.WithTimeout(r.ctx, r.client.timeout)
	defer cancel()

	var lastErr error
	for retry := 0; retry <= r.client.maxRetries; retry++ {
		// Используем r.response как контейнер для ответа
		err := r.client.conn.Invoke(ctx, r.method, r.body, r.response)

		if err == nil {
			return r.response, nil
		}

		st, ok := status.FromError(err)
		if !ok {
			return nil, fmt.Errorf("неизвестная ошибка: %w", err)
		}

		if isRetryable(st.Code()) && retry < r.client.maxRetries {
			time.Sleep(calculateBackoff(retry))
			continue
		}

		lastErr = fmt.Errorf("ошибка gRPC [%s]: %w", st.Code(), err)
		break
	}

	return nil, lastErr
}

// isRetryable проверяет, можно ли повторить запрос при данной ошибке
func isRetryable(code codes.Code) bool {
	switch code {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

// calculateBackoff вычисляет время ожидания между повторными попытками
func calculateBackoff(retry int) time.Duration {
	backoff := time.Duration(1<<uint(retry)) * time.Second
	if backoff > 30*time.Second {
		backoff = 30 * time.Second
	}
	return backoff
}

// Close закрывает соединение с gRPC сервером
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
