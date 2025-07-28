package config

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port int    `mapstructure:"port"` // `mapstructure` tag 用于 Viper 映射
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AppConfig struct {
	Server       ServerConfig   `mapstructure:"server"`
	Mariadb      DatabaseConfig `mapstructure:"mariadb"`
	Redis        RedisConfig    `mapstructure:"redis"`
	Debug        bool           `mapstructure:"debug"`
	JwtSecretKey string         `mapstructure:"jwtsecretkey"`
}

var Config AppConfig

func init() {

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b) // 得到当前文件所在的目录 (例如 my_project/cmd/api)
	// 向上回溯到项目根目录
	// 假设 go.mod 文件总是在项目根目录
	// 你可能需要根据你的实际目录结构调整这个回溯逻辑
	// 例如：如果 main.go 在 cmd/api 下，那么根目录在它上面两层
	projectRoot := filepath.Join(basepath, "..", "..") // 向上两层
	// 如果 main.go 就在根目录下，那就是 projectRoot := basepath
	// 构造 config 目录的绝对路径
	configPath := filepath.Join(projectRoot, "config")
	// 设置配置文件的名称 (无扩展名)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath) // 假设配置文件可能在 config 目录下
	viper.AddConfigPath("./")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 文件未找到错误
			log.Fatalf("配置文件未找到。请确保 'config.yaml' 位于 '%s' 或其他配置路径下。错误: %s", configPath, err)
		} else {
			// 其他读取错误 (例如文件权限，格式错误)
			log.Fatalf("读取配置文件时发生错误。路径: '%s'。错误: %s", configPath, err)
		}
	}

	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatalf("Failed to unmarshal config: %s", err)
	}

	fmt.Printf("Configuration loaded: %+v\n", Config)
	fmt.Println(viper.AllSettings())
	// 根据配置设置 Gin 模式
	if !Config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
}
