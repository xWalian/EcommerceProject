package main

import (
	"context"
	"fmt"
	_ "github.com/lib/pq"
	main4 "github.com/xWalian/EcommerceProject/microservices/logs/server"
	main3 "github.com/xWalian/EcommerceProject/microservices/orders/server"
	main2 "github.com/xWalian/EcommerceProject/microservices/products/server"
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
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
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

	lis, err := net.Listen("tcp", ":50057")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	logConn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer logConn.Close()
	ordersConn, err := grpc.Dial("localhost:50056", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer logConn.Close()
	productsConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer logConn.Close()

	logClient := main4.NewLoggingServiceClient(logConn)
	productsService := main2.NewProductsServiceClient(productsConn)
	ordersService := main3.NewOrdersServiceClient(ordersConn)
	paymentsService := NewServer(client, logClient, ordersService, productsService)
	s := grpc.NewServer()
	RegisterPaymentsServiceServer(s, paymentsService)

	reflection.Register(s)
	log.Printf("Server started listening on %s", lis.Addr().String())

	// Starting gRPC server

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
