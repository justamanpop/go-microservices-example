package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"github.com/justamanpop/microservices_logger/data"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoUrl = "mongodb://mongo:27017"
	gRpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	client = mongoClient
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	app := Config{
		Models: data.New(client),
	}

	err = rpc.Register(new(RpcServer))
	go app.rpcListen()

	go app.grpcListen()

	log.Println("Starting service on port", webPort)
	app.serve()
}

func (cfg *Config) serve() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: cfg.routes(),
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func (cfg *Config) rpcListen() error {
	log.Println("Starting RPC server on port", rpcPort)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			log.Println("Something went wrong in accepting connection", err)
			continue
		}
		go rpc.ServeConn(rpcConn)
	}
}

func connectToMongo() (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(mongoUrl)
	clientOpts.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	c, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Println("Error connecting:", err)
		return nil, err
	}
	log.Println("Connected to mongo")
	return c, nil
}
