package api

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/bondzai/logger/internal/mongodb"
	pb "github.com/bondzai/logger/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoggerServer struct {
	pb.UnimplementedAlertLoggerServer
	Database mongodb.MongoDB
}

type Task struct {
	ID           int         `bson:"task_id" json:"task_id"`
	Organization string      `bson:"organization" json:"organization"`
	ProjectID    int         `bson:"project_id" json:"project_id"`
	Type         pb.TaskType `bson:"type" json:"type"`
	Name         string      `bson:"task_name" json:"task_name"`
	Interval     int64       `bson:"interval" json:"interval"`
	CronExpr     []string    `bson:"task_cron_expression" json:"task_cron_expression"`
	Disabled     bool        `bson:"disabled" json:"disabled"`
	TimeStamp    string      `bson:"timestamp" json:"timestamp"`
}

func (s *LoggerServer) HealthCheck(ctx context.Context, request *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	message := fmt.Sprintf("Health check successful.%s", request)
	return &pb.HealthCheckResponse{Status: message}, nil
}

func StartGRPCServer(database mongodb.MongoDB) error {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterAlertLoggerServer(server, &LoggerServer{Database: database})

	log.Println("gRPC server listening on :50051")
	return server.Serve(listener)
}

func (s *LoggerServer) GetTasks(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	results, err := s.Database.FindLatestDocuments("logs") // Replace with your actual collection name
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get latest tasks: %v", err)
	}

	var tasks []*pb.Task
	for _, result := range results {
		document, ok := result.(primitive.D)
		if !ok {
			log.Printf("Unexpected document format: %+v", result)
			continue
		}

		// Convert primitive.D to byte slice
		data, err := bson.Marshal(document)
		if err != nil {
			log.Printf("Failed to marshal document: %v", err)
			continue
		}

		// Use bson.Unmarshal to decode the byte slice into a Task struct
		var task Task
		err = bson.Unmarshal(data, &task)
		if err != nil {
			log.Printf("Failed to unmarshal document: %v", err)
			continue
		}

		// Convert the Task struct to the gRPC Task message
		pbTask := &pb.Task{
			Id:           int32(task.ID),
			Organization: task.Organization,
			ProjectId:    int32(task.ProjectID),
			Type:         task.Type,
			Name:         task.Name,
			Interval:     task.Interval,
			CronExpr:     task.CronExpr,
			Disabled:     task.Disabled,
		}

		tasks = append(tasks, pbTask)
	}

	return &pb.TaskResponse{Tasks: tasks}, nil
}
