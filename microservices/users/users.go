package users

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	db *mongo.Client
}

func (s *Server) mustEmbedUnimplementedUsersServiceServer() {
}

func (s *Server) GetUser(ctx context.Context, req *GetUserRequest) (*User, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("users")
	var user User
	err := collection.FindOne(ctx, bson.M{"id": req.Id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		return nil, err
	}
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
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		return nil, err
	}
	return &User{
		Id:      req.GetId(),
		Address: req.GetAddress(),
		Phone:   req.GetPhone(),
	}, nil
}

func NewServer(db *mongo.Client) *Server {
	return &Server{db: db}
}
