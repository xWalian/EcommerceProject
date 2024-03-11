package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	main2 "github.com/xWalian/EcommerceProject/microservices/auth/server"
	main3 "github.com/xWalian/EcommerceProject/microservices/logs/server"
	users "github.com/xWalian/EcommerceProject/microservices/users/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	connStr := "user=postgres dbname=db_users password=password host=localhost port=5432 sslmode=disable"
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

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	authConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer authConn.Close()
	logConn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer logConn.Close()
	authClient := main2.NewAuthServiceClient(authConn)
	logClient := main3.NewLoggingServiceClient(logConn)
	s := grpc.NewServer()
	usersService := users.NewServer(db, logClient, authClient)
	users.RegisterUsersServiceServer(s, usersService)
	reflection.Register(s)
	log.Printf("Server started listening on %s", lis.Addr().String())

	// Starting gRPC server

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
