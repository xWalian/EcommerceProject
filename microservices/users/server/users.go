package users

import (
	"context"
	"database/sql"
	auth "github.com/xWalian/EcommerceProject/microservices/auth/pb"
	logs "github.com/xWalian/EcommerceProject/microservices/logging/pb"
	pb "github.com/xWalian/EcommerceProject/microservices/users/pb"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Server struct {
	pb.UnimplementedUsersServiceServer
	db   *sql.DB
	logs logs.LoggingServiceClient
	auth auth.AuthServiceClient
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	query := "SELECT id, password, role FROM users WHERE username = $1"
	var hashedPassword, role string
	var userID int64
	err := s.db.QueryRowContext(ctx, query, req.GetUsername()).Scan(&userID, &hashedPassword, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logs.CreateLog(
				ctx, &logs.CreateLogRequest{
					Service:   "authservice",
					Level:     "WARNING",
					Message:   req.GetUsername() + " Invalid username or password",
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				},
			)
			return nil, status.Errorf(codes.NotFound, "Invalid username or password")
		}
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "authservice",
				Level:     "ERROR",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.GetPassword()))
	if err != nil {
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "authservice",
				Level:     "WARNING",
				Message:   req.GetUsername() + " Invalid username or password",
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, status.Errorf(codes.NotFound, "Invalid username or password")
	}

	tokens, err := s.auth.GenerateToken(
		ctx, &auth.GenerateTokenRequest{
			Id:       userID,
			Role:     role,
			Username: req.GetUsername(),
		},
	)
	if err != nil {
		return nil, err
	}
	s.logs.CreateLog(
		ctx, &logs.CreateLogRequest{
			Service:   "authservice",
			Level:     "INFO",
			Message:   " User logged successfully",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	return &pb.LoginResponse{Token: tokens.Token, Refreshtoken: tokens.RefreshToken}, nil
}

func (s *Server) mustEmbedUnimplementedUsersServiceServer() {
}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	query := "SELECT * FROM users WHERE id = $1"
	row := s.db.QueryRowContext(ctx, query, req.GetId())
	var user pb.User
	err := row.Scan(&user.Id, &user.Username, &user.Password, &user.Email, &user.Address, &user.Phone, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logs.CreateLog(
				ctx, &logs.CreateLogRequest{
					Service:   "userservice",
					Level:     "ERROR",
					Message:   " User not found",
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				},
			)
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "userservice",
				Level:     "ERROR",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, err
	}
	s.logs.CreateLog(
		ctx, &logs.CreateLogRequest{
			Service:   "userservice",
			Level:     "INFO",
			Message:   " Success of finding user",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	return &user, nil
}

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	checkQuery := "SELECT id FROM users WHERE email = $1 OR username = $2"
	var existingUserID int64
	err := s.db.QueryRowContext(ctx, checkQuery, req.GetEmail(), req.GetUsername()).Scan(&existingUserID)
	if err == nil {
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "usersservice",
				Level:     "WARNING",
				Message:   "User with email or username already exists",
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, status.Errorf(codes.AlreadyExists, "User with email or username already exists")
	} else if err != sql.ErrNoRows {
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "usersservice",
				Level:     "ERROR",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "usersservice",
				Level:     "ERROR",
				Message:   err.Error() + " Failed to hash password",
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, status.Errorf(codes.Internal, "Failed to hash password: %v", err)
	}

	insertQuery := "INSERT INTO users (username, email, password, role, address, phone) VALUES ($1, $2, $3, $4, '', '') RETURNING id"
	var userID int64
	err = s.db.QueryRowContext(
		ctx, insertQuery, req.GetUsername(), req.GetEmail(), string(hashedPassword), "user",
	).Scan(&userID)
	if err != nil {
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "usersservice",
				Level:     "ERROR",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, err
	}

	token, err := s.auth.GenerateToken(
		ctx, &auth.GenerateTokenRequest{
			Id:       userID,
			Role:     "user",
			Username: req.Username,
		},
	)
	if err != nil {
		return nil, err
	}

	_, err = s.logs.CreateLog(
		ctx, &logs.CreateLogRequest{
			Service:   "usersservice",
			Level:     "INFO",
			Message:   "User added successfully",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{Token: token.Token, Refreshtoken: token.RefreshToken}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	query := "UPDATE users SET address = $1, phone = $2 WHERE id = $3"
	_, err := s.db.ExecContext(ctx, query, req.GetAddress(), req.GetPhone(), req.GetId())
	if err != nil {
		if err == sql.ErrNoRows {
			s.logs.CreateLog(
				ctx, &logs.CreateLogRequest{
					Service:   "userservice",
					Level:     "ERROR",
					Message:   string(req.GetId()) + " User not found",
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				},
			)
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		s.logs.CreateLog(
			ctx, &logs.CreateLogRequest{
				Service:   "userservice",
				Level:     "ERROR",
				Message:   err.Error(),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		)
		return nil, err
	}
	s.logs.CreateLog(
		ctx, &logs.CreateLogRequest{
			Service:   "userservice",
			Level:     "INFO",
			Message:   string(req.GetId()) + " User found",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	return &pb.User{
		Address: req.GetAddress(),
		Phone:   req.GetPhone(),
	}, nil
}

func NewServer(db *sql.DB, logs logs.LoggingServiceClient, auth auth.AuthServiceClient) *Server {
	return &Server{db: db, logs: logs, auth: auth}
}
