package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/justamanpop/microservices_logger/data"
	"github.com/justamanpop/microservices_logger/logs"
	"google.golang.org/grpc"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		res := &logs.LogResponse{Result: "failed"}
		return res, err
	}
	res := &logs.LogResponse{Result: "logged"}

	myEntry := data.LogEntry{}
	data, err := myEntry.All()
	for index, entry := range data {
		log.Printf("all the data at index %d is %+v", index, *entry)
	}

	return res, nil
}

func (cfg *Config) grpcListen() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for grpc %v", err)
	}

	server := grpc.NewServer()
	logs.RegisterLogServiceServer(server, &LogServer{Models: cfg.Models})
	log.Printf("Listening to grpc on port %s\n", gRpcPort)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to listen for grpc: %v", err)
	}
}
