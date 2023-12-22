package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	pb "github.com/bondzai/logger/proto"

	"github.com/bondzai/logger/internal/mongodb"
	"github.com/bondzai/logger/internal/rabbitmq"
	"google.golang.org/grpc"
)

type GreeterServer struct {
	pb.UnimplementedGreeterServer
}

func (s *GreeterServer) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloResponse, error) {
	message := fmt.Sprintf("Hello, %s!", request.Name)
	return &pb.HelloResponse{Message: message}, nil
}

func init() {
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	mongodb.InitMongoDB()
}

func main() {
	defer mongodb.CloseMongoDB()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start gRPC server
	go func() {
		defer wg.Done()
		err := startGRPCServer()
		if err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// RabbitMQ consumer setup
	rabbitMQConsumer, err := rabbitmq.NewConsumer("amqp://guest:guest@localhost:5672/", "log")
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ consumer: %v", err)
	}

	go func() {
		defer wg.Done()
		<-signals
		log.Println("Received termination signal. Stopping consumer...")
		rabbitMQConsumer.Stop()
	}()

	wg.Add(2)

	// Start RabbitMQ consumer
	go func() {
		defer wg.Done()
		rabbitMQConsumer.Start(processMessage)
	}()

	log.Printf("Consumer and gRPC server started. To exit, press CTRL+C")
	wg.Wait()
}

func processMessage(message map[string]interface{}) bool {
	log.Printf("Received message: %+v", message)

	err := mongodb.InsertDocument("logs", message)
	if err != nil {
		log.Printf("Failed to insert document into MongoDB: %v", err)
		return false
	}

	log.Printf("Message processed and inserted into MongoDB")
	return true
}

func startGRPCServer() error {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, &GreeterServer{})

	log.Println("gRPC server listening on :50051")
	return server.Serve(listener)
}
