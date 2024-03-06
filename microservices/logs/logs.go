package logs

import "go.mongodb.org/mongo-driver/mongo"

type Server struct {
	db *mongo.Client
}

func (s *Server) GetLogsRequest() {

}

func NewServer(db *mongo.Client) *Server {
	return &Server{db: db}
}
