package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	pb "github.com/xWalian/EcommerceProject/microservices/auth/pb"
	auth "github.com/xWalian/EcommerceProject/microservices/auth/server"
	logs "github.com/xWalian/EcommerceProject/microservices/logs/server"
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

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	logConn, err := grpc.Dial("172.17.0.1:50054", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer logConn.Close()
	logClient := logs.NewLoggingServiceClient(logConn)
	s := grpc.NewServer()
	authService := auth.NewServer(db, logClient)
	pb.RegisterAuthServiceServer(s, authService)
	reflection.Register(s)
	log.Printf("Server started listening on %s", lis.Addr().String())

	// Starting gRPC server

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
