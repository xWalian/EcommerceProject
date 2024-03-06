package orders

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	db *mongo.Client
}

func (s *Server) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("orders")
	orderID := uuid.New().String()
	order := &Order{
		Id:         orderID,
		UserId:     req.GetUserId(),
		ProductsId: req.GetProductIds(),
		Status:     "created",
	}
	_, err := collection.InsertOne(ctx, order)
	if err != nil {
		return nil, err
	}
	return order, nil
}
func (s *Server) GetOrder(ctx context.Context, req *GetOrderRequest) (*Order, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("orders")
	var order Order
	err := collection.FindOne(ctx, bson.M{"id": req.Id}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "Order not found")
		}
		return nil, err
	}
	return &order, nil
}
func (s *Server) GetUserOrders(ctx context.Context, req *GetUserOrdersRequest) (*GetUserOrdersResponse, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("orders")
	cursor, err := collection.Find(ctx, bson.M{"userId": req.UserId})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user orders: %v", err)
	}
	defer cursor.Close(ctx)

	var orders []*Order
	for cursor.Next(ctx) {
		var order Order
		if err := cursor.Decode(&order); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to decode order: %v", err)
		}
		orders = append(orders, &order)
	}

	if err := cursor.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "cursor error: %v", err)
	}

	return &GetUserOrdersResponse{Orders: orders}, nil
}

func NewServer(db *mongo.Client) *Server {
	return &Server{db: db}
}
