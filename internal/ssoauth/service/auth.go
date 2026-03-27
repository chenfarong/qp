package service

import (
	"errors"
	"time"

	"github.com/aoyo/qp/internal/ssoauth/model"
	"github.com/aoyo/qp/pkg/db"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务
type AuthService struct {
	db        *db.DB
	jwtSecret string
	expire    int
	dbName    string
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *db.DB, jwtSecret string, expire int, dbName string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: jwtSecret,
		expire:    expire,
		dbName:    dbName,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	Token    string     `json:"token"`
	ExpireAt int64      `json:"expire_at"`
	UserInfo model.User `json:"user_info"`
}

// Register 用户注册
func (s *AuthService) Register(req RegisterRequest) (*TokenResponse, error) {
	collection := s.db.GetCollection(s.dbName, model.User{}.CollectionName())

	// 检查用户名是否已存在
	var existingUser model.User
	if err := collection.FindOne(s.db.Ctx, bson.M{"username": req.Username}).Decode(&existingUser); err == nil {
		return nil, errors.New("username already exists")
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// 检查邮箱是否已存在
	if err := collection.FindOne(s.db.Ctx, bson.M{"email": req.Email}).Decode(&existingUser); err == nil {
		return nil, errors.New("email already exists")
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	now := time.Now()
	user := model.User{
		ID:        primitive.NewObjectID(),
		CreatedAt: now,
		UpdatedAt: now,
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		Nickname:  req.Nickname,
		Status:    1,
	}

	if _, err := collection.InsertOne(s.db.Ctx, user); err != nil {
		return nil, err
	}

	// 生成令牌
	token, expireAt, err := s.generateToken(user.ID.Hex())
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		Token:    token,
		ExpireAt: expireAt,
		UserInfo: user,
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(req LoginRequest) (*TokenResponse, error) {
	collection := s.db.GetCollection(s.dbName, model.User{}.CollectionName())

	// 查找用户
	var user model.User
	if err := collection.FindOne(s.db.Ctx, bson.M{"username": req.Username}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	// 检查密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("user account is disabled")
	}

	// 生成令牌
	token, expireAt, err := s.generateToken(user.ID.Hex())
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		Token:    token,
		ExpireAt: expireAt,
		UserInfo: user,
	}, nil
}

// Claims JWT声明
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// generateToken 生成JWT令牌
func (s *AuthService) generateToken(userID string) (string, int64, error) {
	expireAt := time.Now().Add(time.Hour * time.Duration(s.expire)).Unix()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(expireAt, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expireAt, nil
}

// ValidateToken 验证JWT令牌
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserByID 根据ID获取用户信息
func (s *AuthService) GetUserByID(userID string) (*model.User, error) {
	collection := s.db.GetCollection(s.dbName, model.User{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var user model.User
	if err := collection.FindOne(s.db.Ctx, bson.M{"_id": objectID}).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
