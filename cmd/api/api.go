// api.go
package api

import (
	"fmt"
	"log"

	"github.com/MGJYW/QSLCardManageServer/internal/config"
	"github.com/MGJYW/QSLCardManageServer/internal/controller"
	"github.com/MGJYW/QSLCardManageServer/internal/services"
	"github.com/gin-gonic/gin"
) // 导入 Gin 包

// 全局变量来存储加载的配置

func RunApiService() {
	server := gin.Default()
	regTestApi(server)
	regUserApi(server)
	addr := getAddress()

	log.Printf("服务器将在 %s 上启动...", addr)
	if err := server.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}

	server.Run(addr) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getAddress() string {
	// 定义默认端口
	const defaultPort = 8080

	// 检查端口是否指定，如果未指定或为0，则使用默认端口
	if config.Config.Server.Port == 0 {
		config.Config.Server.Port = defaultPort
		fmt.Printf("警告: 配置文件中未指定端口，使用默认端口 %d。\n", defaultPort)
	}

	// 拼接地址：如果Host为空，则直接使用 ":端口" 的形式
	if config.Config.Server.Host == "" {
		return fmt.Sprintf(":%d", config.Config.Server.Port)
	}

	// Host不为空，则使用 "Host:端口" 的形式
	return fmt.Sprintf("%s:%d", config.Config.Server.Host, config.Config.Server.Port)
}

func regUserApi(server *gin.Engine) {
	userService, err := services.GetUserService()
	if err != nil {
		log.Fatalf("获取用户服务单例失败: %v", err)
	}
	userController := controller.NewUserController(userService)
	userGroup := server.Group("/api/users")
	{
		// userGroup.GET("", userController.GetUsers)
		userGroup.POST("/create", userController.CreateUser)
		userGroup.POST("/login", userController.LoginUser)
		userGroup.GET("/:id", userController.GetUserByID)
	}
}

func regTestApi(server *gin.Engine) {
	server.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}
