package api

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/bondzai/logger/internal/model"
	"github.com/bondzai/logger/internal/mongodb"
	pb "github.com/bondzai/logger/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	protocol     = "tcp"
	port         = ":50051"
	defaultLimit = 1000
)

type LoggerServer struct {
	pb.UnimplementedAlertLoggerServer
	Database *mongodb.MongoDB
}

func StartGRPCServer(database *mongodb.MongoDB) error {
	listener, err := net.Listen(protocol, port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterAlertLoggerServer(server, &LoggerServer{Database: database})

	log.Println("gRPC server listening on", port)
	return server.Serve(listener)
}

func (s *LoggerServer) HealthCheck(ctx context.Context, request *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	message := fmt.Sprintf("Health check successful.%s", request)
	return &pb.HealthCheckResponse{Status: message}, nil
}

func (s *LoggerServer) validateGetLogsRequest(req *pb.TaskRequest) error {
	if req.Organization == "" {
		return fmt.Errorf("organization cannot be empty")
	}
	if req.ProjectId == 0 {
		return fmt.Errorf("project id cannot be empty")
	}
	return nil
}

func (s *LoggerServer) GetLogs(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	if err := s.validateGetLogsRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid request: %v", err)
	}

	query := buildMongoQuery(req)

	findOptions := buildMongoFindOptions(req.Limit)

	results, err := s.Database.FindDocuments("logs", query, findOptions)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get logs: %v", err)
	}

	tasks := convertToProtoTasks(results)
	return &pb.TaskResponse{Tasks: tasks}, nil
}

func buildMongoQuery(req *pb.TaskRequest) bson.D {
	query := bson.D{}

	if req.Organization != "" {
		query = append(query, bson.E{Key: "organization", Value: req.Organization})
	}

	if req.ProjectId != 0 {
		query = append(query, bson.E{Key: "project_id", Value: req.ProjectId})
	}

	return query
}

func buildMongoFindOptions(limit int32) *options.FindOptions {
	findOptions := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	if limit > 0 {
		return findOptions.SetLimit(int64(limit))
	}

	return findOptions.SetLimit(defaultLimit)
}

func convertToProtoTasks(results []interface{}) []*pb.Task {
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
		var task model.Task
		err = bson.Unmarshal(data, &task)
		if err != nil {
			log.Printf("Failed to unmarshal document: %v", err)
			continue
		}

		// Convert the Task struct to the gRPC Task message
		pbTask := &pb.Task{
			Id:           int64(task.ID),
			Organization: task.Organization,
			ProjectId:    int64(task.ProjectID),
			Type:         task.Type,
			Name:         task.Name,
			Interval:     task.Interval,
			CronExpr:     task.CronExpr,
			Disabled:     task.Disabled,
		}

		tasks = append(tasks, pbTask)
	}
	return tasks
}
