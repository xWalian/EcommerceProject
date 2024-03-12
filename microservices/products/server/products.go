package products

import (
	"context"
	"github.com/google/uuid"
	logs "github.com/xWalian/EcommerceProject/microservices/logs/server"
	pb "github.com/xWalian/EcommerceProject/microservices/products/pb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Server struct {
	pb.UnimplementedProductsServiceServer
	db   *mongo.Client
	logs logs.LoggingServiceClient
}

func (s *Server) mustEmbedUnimplementedProductsServiceServer() {
}

func (s *Server) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (
	*pb.GetProductsResponse, error,
) {
	collection := s.db.Database("db_products").Collection("products")
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		_, err := s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "productsservice",
				Level:     "WARNING",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	defer cursor.Close(ctx)
	var products []*pb.Product
	for cursor.Next(ctx) {
		var product pb.Product
		err := cursor.Decode(&product)
		if err != nil {
			s.logs.CreateLog(
				ctx, &logs.CreateLogRequest{
					Service:   "productsservice",
					Level:     "WARNING",
					Message:   err.Error(),
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				},
			)
			return nil, err
		}
		products = append(products, &product)
	}
	if err := cursor.Err(); err != nil {
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "productsservice",
				Level:     "ERROR",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, err
	}
	s.logs.CreateLog(
		ctx, &logs.CreateLogRequest{
			Service:   "productsservice",
			Level:     "INFO",
			Message:   "Products fetched successfully",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	return &pb.GetProductsResponse{Products: products}, nil
}

func (s *Server) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	var product pb.Product
	err := collection.FindOne(ctx, bson.M{"id": req.Id}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.logs.CreateLog(
				ctx, &logs.CreateLogRequest{
					Service:   "productsservice",
					Level:     "WARNING",
					Message:   req.GetId() + "Product not found",
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				},
			)
			return nil, status.Errorf(codes.NotFound, "Product not found")
		}
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "productsservice",
				Level:     "ERROR",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, err
	}
	s.logs.CreateLog(
		ctx, &logs.CreateLogRequest{
			Service:   "productsservice",
			Level:     "INFOR",
			Message:   req.GetId() + "Product fetched successfully",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	return &product, nil
}

func (s *Server) AddProduct(ctx context.Context, req *pb.AddProductRequest) (*pb.Product, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	productID := uuid.New().String
	product := &pb.Product{
		Id:            productID(),
		Name:          req.GetName(),
		Description:   req.GetDescription(),
		Price:         req.GetPrice(),
		StockQuantity: req.GetStockQuantity(),
	}
	_, err := collection.InsertOne(ctx, product)
	if err != nil {
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "productsservice",
				Level:     "ERROR",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, err
	}
	_, err = s.logs.CreateLog(
		ctx, &logs.CreateLogRequest{
			Service:   "productsservice",
			Level:     "INFO",
			Message:   productID() + "Product added successfully",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (s *Server) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	filter := bson.M{"id": req.GetId()}
	update := bson.M{
		"$set": bson.M{
			"name":           req.GetName(),
			"description":    req.GetDescription(),
			"price":          req.GetPrice(),
			"stock_quantity": req.GetStockQuantity(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	// Pobierz zaktualizowany produkt z bazy danych i zwróć go
	var updatedProduct pb.Product
	err = collection.FindOne(ctx, filter).Decode(&updatedProduct)
	if err != nil {
		return nil, err
	}
	s.logs.CreateLog(
		ctx, &logs.CreateLogRequest{
			Service:   "productsservice",
			Level:     "INFO",
			Message:   req.GetId() + "Product added successfully",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	return &updatedProduct, nil
}

func (s *Server) DeleteProduct(
	ctx context.Context, req *pb.DeleteProductRequest,
) (*pb.DeleteProductResponse, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	filter := bson.M{"id": req.GetId()}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	if result.DeletedCount == 0 {
		return &pb.DeleteProductResponse{Success: false}, status.Errorf(codes.NotFound, "Product not found")
	}
	s.logs.CreateLog(
		ctx, &logs.CreateLogRequest{
			Service:   "productsservice",
			Level:     "INFO",
			Message:   req.GetId() + "Product deleted successfully",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	return &pb.DeleteProductResponse{Success: true}, nil
}

func NewServer(db *mongo.Client, logs logs.LoggingServiceClient) *Server {

	return &Server{db: db, logs: logs}
}
