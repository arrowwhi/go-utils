package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arrowwhi/go-utils/grpcclient"
	testproto "github.com/arrowwhi/go-utils/grpcserver/test/proto"
	"google.golang.org/grpc/metadata"
)

func main() {
	// Initialize the gRPC client
	client, err := grpcclient.NewClient(
		"localhost:50051",
		grpcclient.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Create the request using the generated proto message
	req := grpcclient.NewRequest(client, "message_service.users.v1.UsersService/GetStatusInfo", &testproto.Req{
		Input: 3,
	}, &testproto.Resp{}).
		WithMetadata(metadata.Pairs("key", "value")).
		WithContext(context.Background())

	// Execute the request
	response, err := req.Do()
	if err != nil {
		log.Printf("Ошибка запроса: %v", err)
		return
	}

	// Cast the response to the expected type
	resp := response.(*testproto.Resp)
	fmt.Printf("Ответ сервера: %v\n", resp)
}
