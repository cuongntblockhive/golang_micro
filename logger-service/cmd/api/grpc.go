package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"logger/data"
	"logger/logs"
	"net"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	entry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}
	err := l.Models.LogEntry.Insert(entry)

	if err != nil {
		return &logs.LogResponse{Result: "Failed"}, err
	}
	return &logs.LogResponse{Result: "Logged"}, nil
}

func (app *Config) grpcListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("fail to listen for grpc: %v", err)
	}

	s := grpc.NewServer()

	logs.RegisterLogServiceServer(s, &LogServer{
		Models: app.Models,
	})

	log.Printf("gRPC Server started on port %s", grpcPort)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to listen for gRPC %v", err)
	}
}
