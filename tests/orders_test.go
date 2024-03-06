// orders_test.go w folderze tests

package tests

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/xWalian/EcommerceProject/microservices/orders"
)

func TestCreateOrder(t *testing.T) {
	// Utwórz nowego klienta MongoDB w pamięci.
	db, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	assert.NoError(t, err)
	defer db.Disconnect(context.Background())

	// Utwórz nowy serwer zamówień z podłączoną bazą danych MongoDB.
	server := orders.NewServer(db)

	// Wywołaj funkcję CreateOrder.
	order, err := server.CreateOrder(context.Background(), &orders.CreateOrderRequest{
		UserId:     "user123",
		ProductIds: []string{"product1", "product2"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, order)
}

func TestGetOrder(t *testing.T) {
	// Utwórz nowego klienta MongoDB w pamięci.
	db, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	assert.NoError(t, err)
	defer db.Disconnect(context.Background())

	// Utwórz nowy serwer zamówień z podłączoną bazą danych MongoDB.
	server := orders.NewServer(db)

	// Wstaw przykładowe dane do bazy danych.
	_, err = db.Database("test_database").Collection("orders").InsertOne(context.Background(), bson.M{
		"id":     "order123",
		"userId": "user123",
		"productsId": []string{
			"product1",
			"product2",
		},
		"status": "created",
	})
	assert.NoError(t, err)

	// Wywołaj funkcję GetOrder.
	order, err := server.GetOrder(context.Background(), &orders.GetOrderRequest{Id: "order123"})
	assert.NoError(t, err)
	assert.NotNil(t, order)
}

func TestGetUserOrders(t *testing.T) {
	// Utwórz nowego klienta MongoDB w pamięci.
	db, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	assert.NoError(t, err)
	defer db.Disconnect(context.Background())

	// Utwórz nowy serwer zamówień z podłączoną bazą danych MongoDB.
	server := orders.NewServer(db)

	// Wstaw przykładowe dane do bazy danych.
	_, err = db.Database("test_database").Collection("orders").InsertOne(context.Background(), bson.M{
		"id":     "order123",
		"userId": "user123",
		"productsId": []string{
			"product1",
			"product2",
		},
		"status": "created",
	})
	assert.NoError(t, err)

	// Wywołaj funkcję GetUserOrders.
	ordersResponse, err := server.GetUserOrders(context.Background(), &orders.GetUserOrdersRequest{UserId: "user123"})
	assert.NoError(t, err)
	assert.NotNil(t, ordersResponse)
	assert.Len(t, ordersResponse.Orders, 1)
}
