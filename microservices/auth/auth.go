package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/xWalian/EcommerceProject/microservices/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

const (
	accessTokenDuration  = time.Minute * 15
	refreshTokenDuration = time.Hour * 24 * 7
	secretKey            = "aog23noq23jm1ofn234otnqASDF2wfe2"
)

type Server struct {
	db *mongo.Client
}

func (s *Server) VerifyToken(ctx context.Context, request *VerifyTokenRequest) (*VerifyTokenResponse, error) {
	claims, err := ValidateToken(ctx, request.Token)
	if err != nil {
		return nil, err
	}
	userID := claims["id"].(string)
	collection := s.db.Database("db_ecommerce_mongo").Collection("users")
	var user bson.M
	err = collection.FindOne(ctx, bson.M{"id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		return nil, err
	}

	return &VerifyTokenResponse{Valid: true}, nil
}

func (s *Server) mustEmbedUnimplementedAuthServiceServer() {
}

func (s *Server) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("users")
	userId := uuid.New().String()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)

	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}
	user := &users.User{
		Id:       userId,
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Password: string(hashedPassword),
		Address:  "",
		Phone:    "",
		Role:     "user",
	}
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	token, refreshtoken, _ := GenerateToken(user.Id, user.Role, user.Username)

	return &RegisterResponse{Token: token, Refreshtoken: refreshtoken}, nil

}

func (s *Server) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	collection := s.db.Database("db_ecommerce_mongo").Collection("users")
	var user bson.M
	err := collection.FindOne(ctx, bson.M{"username": req.GetUsername()}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	hashedPassword := user["password"].(string)
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.GetPassword()))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	token, refreshtoken, _ := GenerateToken(user["id"].(string), user["role"].(string), user["password"].(string))
	return &LoginResponse{Token: token, Refreshtoken: refreshtoken}, nil
}

func GenerateToken(userid string, userrole string, username string) (string, string, error) {
	accessTokenClaims := jwt.MapClaims{
		"id":       userid,
		"role":     userrole,
		"username": username,
		"exp":      time.Now().Add(accessTokenDuration).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	refreshTokenClaims := jwt.MapClaims{
		"id":  userid,
		"exp": time.Now().Add(refreshTokenDuration).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
func ValidateToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
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
	collection := s.db.Database("db_ecommerce_mongo").Collection("users")
	var user bson.M
	err = collection.FindOne(ctx, bson.M{"id": claims["id"].(string)}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("there is no user connected to this refresh token")
		}
		return nil, err
	}

	userID := claims["id"].(string)
	userrole := user["role"].(string)
	username := user["username"].(string)
	accessToken, _, err := GenerateToken(userID, userrole, username)
	if err != nil {
		return nil, err
	}
	return &RefreshTokenResponse{Token: accessToken}, nil
}

func NewServer(db *mongo.Client) *Server {
	return &Server{db: db}
}
