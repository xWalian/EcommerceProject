package logs

import (
	"context"
	"database/sql"
	"fmt"
	logs "github.com/xWalian/EcommerceProject/microservices/logs/pb"
)

type Server struct {
	logs.UnimplementedLoggingServiceServer
	db *sql.DB
}

func (s *Server) mustEmbedUnimplementedLoggingServiceServer() {

}

func (s *Server) GetLogs(ctx context.Context, req *logs.GetLogsRequest) (*logs.GetLogsResponse, error) {

	query := "SELECT id, service, level, message, timestamp FROM logs WHERE 1=1"
	args := make([]interface{}, 0)

	if req.Service != "" {
		query += " AND service = $1"
		args = append(args, req.Service)
	}
	if req.Level != "" {
		query += " AND level = $2"
		args = append(args, req.Level)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var logers []*logs.Log
	for rows.Next() {
		var log logs.Log
		err := rows.Scan(&log.Id, &log.Service, &log.Level, &log.Message, &log.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		logers = append(logers, &log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return &logs.GetLogsResponse{Logs: logers}, nil
}

func (s *Server) CreateLog(ctx context.Context, request *logs.CreateLogRequest) (*logs.Log, error) {

	query := "INSERT INTO logs (service, level, message, timestamp) VALUES ($1, $2, $3, $4) RETURNING id"
	var id int

	err := s.db.QueryRowContext(
		ctx, query, request.Service, request.Level, request.Message, request.Timestamp,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to insert log: %v", err)
	}

	return &logs.Log{
		Id:        string(rune(id)),
		Service:   request.Service,
		Level:     request.Level,
		Message:   request.Message,
		Timestamp: request.Timestamp,
	}, nil
}

func (s *Server) DeleteLog(ctx context.Context, req *logs.DeleteLogRequest) (*logs.DeleteLogResponse, error) {
	query := "DELETE FROM logs WHERE id = $1"

	result, err := s.db.ExecContext(ctx, query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to execute DELETE query: %v", err)
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %v", err)
	}

	if numRows != 1 {
		return nil, fmt.Errorf("expected to delete 1 row, but deleted %d rows", numRows)
	}

	return &logs.DeleteLogResponse{Success: true}, nil
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}
