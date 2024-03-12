package payments

import (
	"context"
	logs "github.com/xWalian/EcommerceProject/microservices/logs/server"
	orders "github.com/xWalian/EcommerceProject/microservices/orders/server"
	pb "github.com/xWalian/EcommerceProject/microservices/payments/pb"
	products "github.com/xWalian/EcommerceProject/microservices/products/server"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type Server struct {
	pb.UnimplementedPaymentsServiceServer
	db       *mongo.Client
	logs     logs.LoggingServiceClient
	products products.ProductsServiceClient
	orders   orders.OrdersServiceClient
}

func (s *Server) mustEmbedUnimplementedPaymentsServiceServer() {
}

func (s *Server) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (
	*pb.PaymentResponse, error,
) {
	sum, err := s.CheckPrice(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	log.Printf("%v", sum)
	return nil, nil
}

func (s *Server) CheckPrice(ctx context.Context, orderId string) (float64, error) {
	collection := s.db.Database("db_products").Collection("products")
	ordercollection := s.db.Database("db_orders").Collection("orders")

	var order orders.Order
	err := ordercollection.FindOne(ctx, bson.M{"products_id": orderId}).Decode(&order)
	if err != nil {
		return 0, err
	}

	total := 0.0
	for _, productID := range order.ProductsId {
		var product products.Product
		err := collection.FindOne(ctx, bson.M{"id": productID}).Decode(&product)
		if err != nil {
			return 0, err
		}
		total += float64(product.Price)
	}

	return total, nil
}

func NewServer(
	db *mongo.Client, logs logs.LoggingServiceClient, orders orders.OrdersServiceClient,
	products products.ProductsServiceClient,
) *Server {
	return &Server{db: db, logs: logs, orders: orders, products: products}
}
