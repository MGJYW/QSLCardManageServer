// internal/services/mariadb.go
package services

import (
	"fmt"
	"log"
	"sync"

	"github.com/Hope-Cruiser-Psy-Volunteer-Alliance/HOMB/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GORMDBService 是基于 GORM 的数据库服务接口
type GORMDBService interface {
	GetDB() *gorm.DB // 返回 GORM 的 DB 实例
	Close() error    // 关闭底层 SQL 连接池

	// GORM 的 AutoMigrate 通常在启动时运行，用于自动创建或更新表结构
	AutoMigrate(models ...interface{}) error
}

// gormDBServiceImpl 是 GORMDBService 接口的实现
type gormDBServiceImpl struct {
	db    *gorm.DB
	sqlDB *gorm.DB // 用于存储 GORM 内部的 *sql.DB，以便关闭
}

// 声明单例实例和 sync.Once
var (
	gormDBSingleton GORMDBService // 修改变量名以反映 GORM
	once            sync.Once
)

// GetGORMDBService 获取 GORM 数据库服务的单例实例
func GetGORMDBService() (GORMDBService, error) {
	var initErr error // 声明一个局部错误变量，以便在闭包中捕获
	once.Do(func() {
		dbconfig := config.Config.Mariadb // 假设 config.Config.Database 已经正确加载了数据库配置

		// 构造连接字符串 (确保 dbconfig.Host, dbconfig.Port, dbconfig.Name, User, Password 存在)
		connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbconfig.User,
			dbconfig.Password,
			dbconfig.Host,
			dbconfig.Port,
			dbconfig.Name,
		)

		// 检查连接字符串是否为空
		if connStr == "" {
			initErr = fmt.Errorf("MariaDB 连接字符串为空，请检查配置文件")
			return // 提前返回，不要继续初始化
		}

		// GORM 连接选项
		gormConfig := &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info), // 在生产环境可以设置为 Silent
		}

		// 使用 GORM 的 mysql 驱动打开数据库连接
		db, openErr := gorm.Open(mysql.Open(connStr), gormConfig)
		if openErr != nil {
			initErr = fmt.Errorf("无法通过 GORM 连接到 MariaDB: %w", openErr)
			return
		}

		// 获取底层 *sql.DB 实例，用于设置连接池参数和 Ping
		sqlDB, getSqlErr := db.DB()
		if getSqlErr != nil {
			initErr = fmt.Errorf("无法获取 GORM 底层 SQL DB: %w", getSqlErr)
			return
		}

		// 设置连接池参数
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(0)

		// 尝试 Ping 数据库
		if pingErr := sqlDB.Ping(); pingErr != nil {
			sqlDB.Close() // ping 失败后立即关闭连接
			initErr = fmt.Errorf("MariaDB 连接失败 (通过 GORM): %w", pingErr)
			return
		}

		log.Println("MariaDB (GORM) 连接成功！")
		// 将 GORM 的 DB 实例赋值给单例变量
		gormDBSingleton = &gormDBServiceImpl{db: db, sqlDB: db} // sqlDB 也指向 GORM 的 db
	})

	// 返回捕获到的错误和单例实例
	return gormDBSingleton, initErr
}

// GetDB 返回底层的 *sql.DB 实例
func (s *gormDBServiceImpl) GetDB() *gorm.DB {
	return s.db
}

// Close 关闭底层 SQL 连接池
func (s *gormDBServiceImpl) Close() error {
	if s.sqlDB != nil { // 使用 sqlDB 字段来获取底层连接
		sqlDB, err := s.sqlDB.DB() // 再次获取 GORM 底层 SQL DB，确保是活动的
		if err != nil {
			return fmt.Errorf("获取 GORM 底层 SQL DB 失败: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}

// AutoMigrate 用于自动迁移数据库表结构
func (s *gormDBServiceImpl) AutoMigrate(models ...interface{}) error {
	log.Println("执行数据库自动迁移...")
	return s.db.AutoMigrate(models...)
}
