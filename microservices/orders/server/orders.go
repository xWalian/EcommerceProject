package orders

import (
	"context"
	"errors"
	"github.com/google/uuid"
	logs "github.com/xWalian/EcommerceProject/microservices/logs/server"
	pb "github.com/xWalian/EcommerceProject/microservices/orders/pb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Server struct {
	pb.UnimplementedOrdersServiceServer
	db   *mongo.Client
	logs logs.LoggingServiceClient
}

func (s *Server) mustEmbedUnimplementedOrdersServiceServer() {

}
func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.Order, error) {
	collection := s.db.Database("db_orders").Collection("pb")
	for _, productID := range req.GetProductIds() {
		exists, err := s.productExists(ctx, productID)
		if err != nil {
			s.logs.CreateLog(
				ctx, &logs.CreateLogRequest{
					Service:   "ordersservice",
					Level:     "ERROR",
					Message:   err.Error(),
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				},
			)
			return nil, err
		}
		if !exists {
			s.logs.CreateLog(
				ctx, &logs.CreateLogRequest{
					Service:   "ordersservice",
					Level:     "WARNING",
					Message:   "product with ID " + productID + " does not exist",
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				},
			)
			return nil, errors.New("product with ID " + productID + " does not exist")
		}
	}
	exists, err := s.userExist(ctx, req.UserId)
	if err != nil {
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "ordersservice",
				Level:     "ERROR",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, err
	}
	if !exists {
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "ordersservice",
				Level:     "WARNING",
				Message:   "user with ID " + req.UserId + " does not exist",
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, errors.New("user with ID " + req.UserId + " does not exist")
	}

	orderID := uuid.New().String()
	order := &pb.Order{
		Id:         orderID,
		UserId:     req.GetUserId(),
		ProductsId: req.GetProductIds(),
		Status:     "created",
	}
	_, err = collection.InsertOne(ctx, order)
	if err != nil {
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "ordersservice",
				Level:     "ERROR",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, err
	}
	for _, productID := range req.GetProductIds() {
		if err := s.decreaseProductQuantity(ctx, productID); err != nil {
			s.logs.CreateLog(
				ctx, &logs.CreateLogRequest{
					Service:   "ordersservice",
					Level:     "ERROR",
					Message:   err.Error(),
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				},
			)
			return nil, err
		}
	}
	s.logs.CreateLog(
		ctx, &logs.CreateLogRequest{
			Service:   "ordersservice",
			Level:     "INFO",
			Message:   orderID + " Order successfully added",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	return order, nil
}
func (s *Server) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.Order, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("pb")
	var order pb.Order
	err := collection.FindOne(ctx, bson.M{"id": req.Id}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.logs.CreateLog(
				ctx, &logs.CreateLogRequest{
					Service:   "ordersservice",
					Level:     "WARNING",
					Message:   req.GetId() + " Order not found",
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				},
			)
			return nil, status.Errorf(codes.NotFound, "Order not found")
		}
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "ordersservice",
				Level:     "ERROR",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, err
	}
	s.logs.CreateLog(
		ctx, &logs.CreateLogRequest{
			Service:   "ordersservice",
			Level:     "INFO",
			Message:   req.GetId() + "Orders fetched successfully",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	return &order, nil
}
func (s *Server) GetUserOrders(ctx context.Context, req *pb.GetUserOrdersRequest) (
	*pb.GetUserOrdersResponse, error,
) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("pb")
	cursor, err := collection.Find(ctx, bson.M{"userId": req.UserId})
	if err != nil {
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "ordersservice",
				Level:     "WARNING",
				Message:   req.UserId + "Failed to get user pb",
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, status.Errorf(codes.Internal, "failed to get user pb: %v", err)
	}
	var orders []*pb.Order
	for cursor.Next(ctx) {
		var order pb.Order
		if err := cursor.Decode(&order); err != nil {
			s.logs.CreateLog(
				ctx, &logs.CreateLogRequest{
					Service:   "ordersservice",
					Level:     "WARNING",
					Message:   req.UserId + "Failed to decode order",
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				},
			)
			return nil, status.Errorf(codes.Internal, "failed to decode order: %v", err)
		}
		orders = append(orders, &order)
	}

	if err := cursor.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "cursor error: %v", err)
	}

	return &pb.GetUserOrdersResponse{Orders: orders}, nil
}

func (s *Server) productExists(ctx context.Context, productID string) (bool, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	var product bson.M
	err := collection.FindOne(ctx, bson.M{"id": productID}).Decode(&product)
	if err != nil {
		return false, err
	}

	quantity, ok := product["quantity"].(int64)
	if !ok {
		return false, errors.New("failed to parse quantity")
	}
	if quantity < 1 {
		return false, errors.New("product quantity must be greater than 0")
	}
	var product2 bson.M
	err = collection.FindOne(ctx, bson.M{"id": productID}).Decode(&product2)
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

func (s *Server) decreaseProductQuantity(ctx context.Context, productID string) error {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")

	filter := bson.M{"id": productID}
	update := bson.M{"$inc": bson.M{"quantity": -1}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func NewServer(db *mongo.Client, logs logs.LoggingServiceClient) *Server {
	return &Server{db: db, logs: logs}
}
