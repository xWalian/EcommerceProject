package main

import (
	"context"
	"fmt"
	_ "github.com/lib/pq"
	main2 "github.com/xWalian/EcommerceProject/microservices/logging/pb"
	pb "github.com/xWalian/EcommerceProject/microservices/products/pb"
	products "github.com/xWalian/EcommerceProject/microservices/products/server"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
)

func main() {
	// Initializing connection with MongoDB

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://172.17.0.1:27017"))
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatalf("failed to disconnect from MongoDB: %v", err)
		}
	}()

	// Checking connection with MongoDB

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("failed to ping MongoDB: %v", err)
	}
	fmt.Println("Connected to MongoDB!")

	// Creating gRPC listener

	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	logConn, err := grpc.Dial("172.17.0.1:50054", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to log service: %v", err)
	}
	defer logConn.Close()
	logClient := main2.NewLoggingServiceClient(logConn)
	s := grpc.NewServer()
	productsService := products.NewServer(client, logClient)
	pb.RegisterProductsServiceServer(s, productsService)
	reflection.Register(s)
	log.Printf("Server started listening on %s", lis.Addr().String())

	// Starting gRPC server

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
