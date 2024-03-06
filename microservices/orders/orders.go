package orders

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	db *mongo.Client
}

func (s *Server) mustEmbedUnimplementedOrdersServiceServer() {

}
func (s *Server) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("orders")
	for _, productID := range req.GetProductIds() {
		exists, err := s.productExists(ctx, productID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("product with ID " + productID + " does not exist")
		}
	}
	exists, err := s.userExist(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("product with ID " + req.UserId + " does not exist")
	}
	orderID := uuid.New().String()
	order := &Order{
		Id:         orderID,
		UserId:     req.GetUserId(),
		ProductsId: req.GetProductIds(),
		Status:     "created",
	}
	_, err = collection.InsertOne(ctx, order)
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

func (s *Server) productExists(ctx context.Context, productID string) (bool, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	var product bson.M
	err := collection.FindOne(ctx, bson.M{"id": productID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Server) userExist(ctx context.Context, userId string) (bool, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	var product bson.M
	err := collection.FindOne(ctx, bson.M{"id": userId}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func NewServer(db *mongo.Client) *Server {
	return &Server{db: db}
}
