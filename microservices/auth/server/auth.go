package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	logs "github.com/xWalian/EcommerceProject/microservices/logs/server"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

const (
	accessTokenDuration  = time.Minute * 15
	refreshTokenDuration = time.Hour * 24 * 7
	secretKey            = "aog23noq23jm1ofn234otnqASDF2wfe2"
)

type Server struct {
	db   *sql.DB
	logs logs.LoggingServiceClient
}

func (s *Server) VerifyToken(ctx context.Context, request *VerifyTokenRequest) (*VerifyTokenResponse, error) {
	claims, err := ValidateToken(ctx, request.Token)
	if err != nil {
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

	userID := claims["id"]
	query := "SELECT id FROM users WHERE id = $1"
	var id string
	err = s.db.QueryRowContext(ctx, query, userID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logs.CreateLog(
				ctx, &logs.CreateLogRequest{
					Service:   "authservice",
					Level:     "WARNING",
					Message:   " User not found",
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				},
			)
			return nil, status.Errorf(codes.NotFound, "User not found")
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

	return &VerifyTokenResponse{Valid: true}, nil
}

func (s *Server) mustEmbedUnimplementedAuthServiceServer() {
}

func (s *Server) GenerateToken(ctx context.Context, req *GenerateTokenRequest) (*GenerateTokenResponse, error) {
	accessTokenClaims := jwt.MapClaims{
		"id":       req.GetId(),
		"role":     req.Role,
		"username": req.GetUsername(),
		"exp":      time.Now().Add(accessTokenDuration).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(secretKey))
	if err != nil {

		return nil, err
	}

	refreshTokenClaims := jwt.MapClaims{
		"id":  req.GetId(),
		"exp": time.Now().Add(refreshTokenDuration).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &GenerateTokenResponse{
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}
func ValidateToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(
		tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		},
	)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
func (s *Server) RefreshToken(ctx context.Context, request *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	claims, err := ValidateToken(ctx, request.GetRefreshtoken())
	if err != nil {
		return nil, err
	}
	query := "SELECT role, username FROM users WHERE id = $1"
	var userrole, username string
	err = s.db.QueryRowContext(ctx, query, claims["id"]).Scan(&userrole, &username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("there is no user connected to this refresh token")
		}
		return nil, err
	}
	fmt.Printf("Type of claims['id']: %T\n", claims["id"])

	idFloat, ok := claims["id"].(float64)
	if !ok {
		return nil, errors.New("failed to convert user ID to float64")
	}
	userID := int64(idFloat)
	accessToken, _ := s.GenerateToken(
		ctx, &GenerateTokenRequest{
			Id:       userID,
			Role:     userrole,
			Username: username,
		},
	)
	if err != nil {
		return nil, err
	}
	return &RefreshTokenResponse{Token: accessToken.Token}, nil
}

func NewServer(db *sql.DB, logs logs.LoggingServiceClient) *Server {
	return &Server{db: db, logs: logs}
}
