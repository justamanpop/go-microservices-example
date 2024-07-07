package main

import (
	"context"
	"log"
	"time"

	"github.com/justamanpop/microservices_logger/data"
)

type RpcServer struct {
}

type RpcPayload struct {
	Name string
	Data string
}

func (r *RpcServer) Log(payload RpcPayload, res *string) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})

	if err != nil {
		log.Println("Error writing to mongo")
		return err
	}
	*res = "Processed req received from RPC: " + payload.Name

	myEntry := data.LogEntry{}
	data, err := myEntry.All()
	for index, entry := range data {
		log.Printf("all the data at index %d is %+v", index, *entry)
	}
	return nil
}
