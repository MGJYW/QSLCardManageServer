// internal/services/user_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Hope-Cruiser-Psy-Volunteer-Alliance/HOMB/internal/config"
	"github.com/Hope-Cruiser-Psy-Volunteer-Alliance/HOMB/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type UserService interface {
	GetUserByText(ctx context.Context, text string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	GenerateToken(ctx context.Context, user *models.User) (string, error)
	VerifyToken(tokenString string) (*models.UserJWT, error)
}

type userServiceImpl struct {
	gormDB  *gorm.DB
	redisDB RedisService
}

var (
	userSingleton UserService
	userOnce      sync.Once
)

var jwtSecret = []byte(config.Config.JwtSecretKey)

func GetUserService() (UserService, error) {
	var initErr error
	userOnce.Do(func() {
		gormDBService, err := GetGORMDBService()
		if err != nil {
			initErr = fmt.Errorf("获取 GORM 数据库服务单例失败: %w", err)
			return
		}
		gormDB := gormDBService.GetDB() // 获取 GORM DB 实例

		redisService, err := GetRedisService()
		if err != nil {
			initErr = fmt.Errorf("获取 Redis 服务单例失败: %w", err)
			return
		}

		userSingleton = &userServiceImpl{gormDB: gormDB, redisDB: redisService}
	})

	return userSingleton, initErr
}

func (s *userServiceImpl) GetUserByText(ctx context.Context, text string) (*models.User, error) {
	log.Printf("UserService: 从 MariaDB (GORM) 获取用户条目: %s...\n", text)
	var user models.User
	result := s.gormDB.WithContext(ctx).Where("email = ?", text).Or("user_id = ?", text).Or("phone = ?", text).Or("name = ?", text).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到用户条目: %s", text)
		}
		return nil, fmt.Errorf("GORM 查询用户失败: %w", result.Error)
	}
	return &user, nil
}

func (s *userServiceImpl) CreateUser(ctx context.Context, user *models.User) error {
	log.Printf("UserService: 创建用户: %+v...\n", user)
	if err := s.gormDB.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("GORM 创建用户失败: %w", err)
	}

	s.redisDB.Del(ctx, "all_users") // 清除所有用户列表缓存

	return nil
}

func (s *userServiceImpl) GenerateToken(ctx context.Context, user *models.User) (string, error) {
	claims := models.UserJWT{
		UserID: user.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // 24小时后过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
			Subject:   user.UserID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *userServiceImpl) VerifyToken(tokenString string) (*models.UserJWT, error) {
	claims := &models.UserJWT{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非法的签名方法: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, fmt.Errorf("JWT 令牌已过期或尚未生效: %w", err)
		}
		return nil, fmt.Errorf("解析 JWT 令牌失败: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("JWT 令牌无效")
	}

	return claims, nil
}
