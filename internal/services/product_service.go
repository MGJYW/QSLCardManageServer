// internal/services/product_service.go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/MGJYW/QSLCardManageServer/internal/models"
	"gorm.io/gorm"
)

// ProductService 接口定义了产品业务操作
type ProductService interface {
	GetAllProducts(ctx context.Context) ([]models.Product, error)
	GetProductByID(ctx context.Context, id uint) (*models.Product, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id uint) error
}

// productServiceImpl 是 ProductService 接口的具体实现
type productServiceImpl struct {
	gormDB  *gorm.DB
	redisDB RedisService // 可能需要缓存产品数据
}

// 产品服务单例相关变量
var (
	productSingleton ProductService
	productOnce      sync.Once
)

// GetProductService 返回产品服务的单例实例
func GetProductService() (ProductService, error) {
	var initErr error
	productOnce.Do(func() {
		// 获取 GORM 数据库服务的单例实例
		gormDBService, err := GetGORMDBService()
		if err != nil {
			initErr = fmt.Errorf("获取 GORM 数据库服务单例失败: %w", err)
			return
		}
		gormDB := gormDBService.GetDB() // 获取 GORM DB 实例

		// 获取 Redis 服务的单例实例
		redisService, err := GetRedisService()
		if err != nil {
			initErr = fmt.Errorf("获取 Redis 服务单例失败: %w", err)
			return
		}

		// 初始化产品服务单例
		productSingleton = &productServiceImpl{
			gormDB:  gormDB,
			redisDB: redisService,
		}
	})

	return productSingleton, initErr
}

// GetAllProducts 获取所有产品（优先从 Redis 缓存，否则从 MariaDB）
func (s *productServiceImpl) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	cacheKey := "all_products"

	// 1. 尝试从 Redis 缓存获取
	cachedProducts, err := s.redisDB.Get(ctx, cacheKey).Bytes()
	if err == nil && len(cachedProducts) > 0 {
		var products []models.Product
		if err := json.Unmarshal(cachedProducts, &products); err == nil {
			log.Println("ProductService: 从 Redis 缓存获取所有产品...")
			return products, nil
		}
	}

	// 2. 缓存中没有，从 MariaDB 获取
	log.Println("ProductService: 从 MariaDB (GORM) 获取所有产品...")
	var products []models.Product
	if err := s.gormDB.WithContext(ctx).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("GORM 查询所有产品失败: %w", err)
	}

	// 3. 将结果存入 Redis 缓存
	productsBytes, err := json.Marshal(products)
	if err == nil {
		s.redisDB.Set(ctx, cacheKey, productsBytes, 5*time.Minute) // 缓存 5 分钟
	} else {
		log.Printf("无法将产品数据序列化为 JSON 进行缓存: %v", err)
	}

	return products, nil
}

// GetProductByID 根据ID获取产品
func (s *productServiceImpl) GetProductByID(ctx context.Context, id uint) (*models.Product, error) {
	log.Printf("ProductService: 从 MariaDB (GORM) 获取产品 ID: %d...\n", id)
	var product models.Product
	result := s.gormDB.WithContext(ctx).First(&product, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到产品 ID: %d", id)
		}
		return nil, fmt.Errorf("GORM 查询产品失败: %w", result.Error)
	}
	return &product, nil
}

// CreateProduct 创建产品
func (s *productServiceImpl) CreateProduct(ctx context.Context, product *models.Product) error {
	log.Printf("ProductService: 创建产品: %+v...\n", product)
	if err := s.gormDB.WithContext(ctx).Create(product).Error; err != nil {
		return fmt.Errorf("GORM 创建产品失败: %w", err)
	}
	s.redisDB.Del(ctx, "all_products") // 清除所有产品列表缓存
	return nil
}

// UpdateProduct 更新产品
func (s *productServiceImpl) UpdateProduct(ctx context.Context, product *models.Product) error {
	log.Printf("ProductService: 更新产品: %+v...\n", product)
	// GORM 会根据结构体中的ID来更新对应的记录
	if err := s.gormDB.WithContext(ctx).Save(product).Error; err != nil {
		return fmt.Errorf("GORM 更新产品失败: %w", err)
	}
	s.redisDB.Del(ctx, "all_products") // 清除所有产品列表缓存
	return nil
}

// DeleteProduct 删除产品（软删除）
func (s *productServiceImpl) DeleteProduct(ctx context.Context, id uint) error {
	log.Printf("ProductService: 删除产品 ID: %d...\n", id)
	// GORM 默认是软删除
	result := s.gormDB.WithContext(ctx).Delete(&models.Product{}, id)
	if result.Error != nil {
		return fmt.Errorf("GORM 删除产品失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("未找到产品 ID: %d 进行删除", id)
	}
	s.redisDB.Del(ctx, "all_products") // 清除所有产品列表缓存
	return nil
}
