// main.go
package main

import (
	"context"
	"fmt"
	"github.com/xWalian/EcommerceProject/microservices/auth"
	"github.com/xWalian/EcommerceProject/microservices/orders"
	"github.com/xWalian/EcommerceProject/microservices/users"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

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

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Making main server and services instances
	s := grpc.NewServer()
	productsService := products.NewServer(client)
	usersService := users.NewServer(client)
	ordersService := orders.NewServer(client)
	authService := auth.NewServer(client)
	products.RegisterProductsServiceServer(s, productsService)
	users.RegisterUsersServiceServer(s, usersService)
	orders.RegisterOrdersServiceServer(s, ordersService)
	auth.RegisterAuthServiceServer(s, authService)
	reflection.Register(s)
	log.Printf("Server started listening on %s", lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
