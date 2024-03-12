package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	logspb "github.com/xWalian/EcommerceProject/microservices/logging/pb"
	logs "github.com/xWalian/EcommerceProject/microservices/logging/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	connStr := "user=postgres dbname=db_logs password=password host=172.17.0.1 port=5432 sslmode=disable"
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

	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	logsService := logs.NewServer(db)
	logspb.RegisterLoggingServiceServer(s, logsService)
	reflection.Register(s)
	log.Printf("Server started listening on %s", lis.Addr().String())

	// Starting gRPC server

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
