package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/xWalian/EcommerceProject/microservices/auth"
	"github.com/xWalian/EcommerceProject/microservices/logs"
	"github.com/xWalian/EcommerceProject/microservices/orders"
	"github.com/xWalian/EcommerceProject/microservices/users"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"

	"github.com/xWalian/EcommerceProject/microservices/products"
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

	//Configuration PostgreSQL database

	connStr := "user=postgres dbname=db_ecommerce_postgresql password=password host=localhost port=5432 sslmode=disable"
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

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Creating instances of main server and services

	s := grpc.NewServer()
	logsService := logs.NewServer(db)
	productsService := products.NewServer(client, logsService)
	usersService := users.NewServer(client, logsService)
	ordersService := orders.NewServer(client, logsService)
	authService := auth.NewServer(client, logsService)

	// Registering services with the main server

	products.RegisterProductsServiceServer(s, productsService)
	users.RegisterUsersServiceServer(s, usersService)
	orders.RegisterOrdersServiceServer(s, ordersService)
	auth.RegisterAuthServiceServer(s, authService)
	logs.RegisterLoggingServiceServer(s, logsService)

	// Registering reflection service for gRPC

	reflection.Register(s)
	log.Printf("Server started listening on %s", lis.Addr().String())

	// Starting gRPC server

	if err := s.Serve(lis); err != nil {
		logsService.CreateLog(ctx,
			&logs.CreateLogRequest{
				Service:   "main",
				Level:     "ERROR",
				Message:   "Failed to serve: " + err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		log.Fatalf("failed to serve: %v", err)
	}
}
