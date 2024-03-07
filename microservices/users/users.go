package users

import (
	"context"
	"github.com/xWalian/EcommerceProject/microservices/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Server struct {
	db   *mongo.Client
	logs *logs.Server
}

func (s *Server) mustEmbedUnimplementedUsersServiceServer() {
}

func (s *Server) GetUser(ctx context.Context, req *GetUserRequest) (*User, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("users")
	var user User
	err := collection.FindOne(ctx, bson.M{"id": req.Id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.logs.CreateLog(ctx, &logs.CreateLogRequest{
				Service:   "userservice",
				Level:     "ERROR",
				Message:   req.GetId() + " User not found",
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			})
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		s.logs.CreateLog(ctx, &logs.CreateLogRequest{
			Service:   "userservice",
			Level:     "ERROR",
			Message:   err.Error(),
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		})
		return nil, err
	}
	s.logs.CreateLog(ctx, &logs.CreateLogRequest{
		Service:   "userservice",
		Level:     "INFO",
		Message:   req.GetId() + " Success of finding user",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	})
	return &user, nil
}
func (s *Server) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*User, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("users")

	update := bson.M{
		"$set": bson.M{
			"address": req.GetAddress(),
			"phone":   req.GetPhone(),
		},
	}
	_, err := collection.UpdateOne(ctx, bson.M{"id": req.GetId()}, update)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.logs.CreateLog(ctx, &logs.CreateLogRequest{
				Service:   "userservice",
				Level:     "ERROR",
				Message:   req.GetId() + " User not found",
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			})
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		s.logs.CreateLog(ctx, &logs.CreateLogRequest{
			Service:   "userservice",
			Level:     "ERROR",
			Message:   err.Error(),
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		})
		return nil, err
	}
	s.logs.CreateLog(ctx, &logs.CreateLogRequest{
		Service:   "userservice",
		Level:     "INFO",
		Message:   req.GetId() + " User found",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	})
	return &User{
		Id:      req.GetId(),
		Address: req.GetAddress(),
		Phone:   req.GetPhone(),
	}, nil
}

func NewServer(db *mongo.Client, logs *logs.Server) *Server {
	return &Server{db: db, logs: logs}
}
