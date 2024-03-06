package products

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

func (s *Server) mustEmbedUnimplementedProductsServiceServer() {
}

func (s *Server) GetProducts(ctx context.Context, req *GetProductsRequest) (*GetProductsResponse, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*Product
	for cursor.Next(ctx) {
		var product Product
		err := cursor.Decode(&product)
		if err != nil {
			return nil, err
		}
		products = append(products, &product)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return &GetProductsResponse{Products: products}, nil
}

func (s *Server) GetProduct(ctx context.Context, req *GetProductRequest) (*Product, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	var product Product
	err := collection.FindOne(ctx, bson.M{"id": req.Id}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "Product not found")
		}
		return nil, err
	}
	return &product, nil
}

func (s *Server) AddProduct(ctx context.Context, req *AddProductRequest) (*Product, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	productID := uuid.New().String
	product := &Product{
		Id:            productID(),
		Name:          req.GetName(),
		Description:   req.GetDescription(),
		Price:         req.GetPrice(),
		StockQuantity: req.GetStockQuantity(),
	}
	_, err := collection.InsertOne(ctx, product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *Server) UpdateProduct(ctx context.Context, req *UpdateProductRequest) (*Product, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	filter := bson.M{"id": req.GetId()}
	update := bson.M{"$set": bson.M{
		"name":           req.GetName(),
		"description":    req.GetDescription(),
		"price":          req.GetPrice(),
		"stock_quantity": req.GetStockQuantity(),
	}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	// Pobierz zaktualizowany produkt z bazy danych i zwróć go
	var updatedProduct Product
	err = collection.FindOne(ctx, filter).Decode(&updatedProduct)
	if err != nil {
		return nil, err
	}

	return &updatedProduct, nil
}

func (s *Server) DeleteProduct(ctx context.Context, req *DeleteProductRequest) (*DeleteProductResponse, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("products")
	filter := bson.M{"id": req.GetId()}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	if result.DeletedCount == 0 {
		return &DeleteProductResponse{Success: false}, status.Errorf(codes.NotFound, "Product not found")
	}
	return &DeleteProductResponse{Success: true}, nil
}

func NewServer(db *mongo.Client) *Server {
	return &Server{db: db}
}
