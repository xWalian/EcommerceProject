package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	auth "github.com/xWalian/EcommerceProject/microservices/auth/server"
	logs "github.com/xWalian/EcommerceProject/microservices/logs/server"
	pb "github.com/xWalian/EcommerceProject/microservices/users/pb"
	users "github.com/xWalian/EcommerceProject/microservices/users/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
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

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	authConn, err := grpc.Dial("172.17.0.1:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer authConn.Close()
	logConn, err := grpc.Dial("172.17.0.1:50054", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to log service: %v", err)
	}
	defer logConn.Close()
	authClient := auth.NewAuthServiceClient(authConn)
	logClient := logs.NewLoggingServiceClient(logConn)
	s := grpc.NewServer()
	usersService := users.NewServer(db, logClient, authClient)
	pb.RegisterUsersServiceServer(s, usersService)
	reflection.Register(s)
	log.Printf("Server started listening on %s", lis.Addr().String())

	// Starting gRPC server

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
