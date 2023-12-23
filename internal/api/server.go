package api

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/bondzai/logger/proto"
	"google.golang.org/grpc"
)

type LoggerServer struct {
	pb.UnimplementedAlertLoggerServer
}

func (s *LoggerServer) HealthCheck(ctx context.Context, request *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	message := fmt.Sprintf("Health check successful.%s", request)
	return &pb.HealthCheckResponse{Status: message}, nil
}

func (s *LoggerServer) GetTasks(ctx context.Context, request *pb.TaskRequest) (*pb.TaskResponse, error) {
	tasks := []*pb.Task{
		{
			Id:           1,
			Organization: "ExampleOrg",
			ProjectId:    123,
			Type:         pb.TaskType_INTERVAL,
			Name:         "Task1",
			Interval:     60,
			CronExpr:     []string{"* * * * *"},
			Disabled:     false,
		},
		{
			Id:           2,
			Organization: "ExampleOrg",
			ProjectId:    456,
			Type:         pb.TaskType_CRON,
			Name:         "Task2",
			CronExpr:     []string{"0 0 * * *"},
			Disabled:     true,
		},
	}

	return &pb.TaskResponse{Tasks: tasks}, nil
}

func StartGRPCServer() error {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterAlertLoggerServer(server, &LoggerServer{})

	log.Println("gRPC server listening on :50051")
	return server.Serve(listener)
}
