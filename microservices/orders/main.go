package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	main2 "github.com/xWalian/EcommerceProject/microservices/logging/pb"
	pb "github.com/xWalian/EcommerceProject/microservices/orders/pb"
	orders "github.com/xWalian/EcommerceProject/microservices/orders/server"
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

	// Initializing connection with PostgreSQL
	connStr := "user=postgres dbname=db_users password=password host=172.17.0.1 port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Checking connection with PostgreSQL

	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	fmt.Println("Connected to PostgreSQL database!")
	// Creating gRPC listener

	lis, err := net.Listen("tcp", ":50056")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	logConn, err := grpc.Dial("172.17.0.1:50054", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer logConn.Close()

	logClient := main2.NewLoggingServiceClient(logConn)
	s := grpc.NewServer()
	ordersService := orders.NewServer(client, logClient, db)
	pb.RegisterOrdersServiceServer(s, ordersService)

	reflection.Register(s)
	log.Printf("Server started listening on %s", lis.Addr().String())

	// Starting gRPC server

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
