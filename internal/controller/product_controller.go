// internal/controllers/product_controller.go
package controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MGJYW/QSLCardManageServer/internal/models"
	"github.com/MGJYW/QSLCardManageServer/internal/services" // 导入 ProductService
	"github.com/gin-gonic/gin"
)

// ProductController 结构体，持有 ProductService 的依赖
type ProductController struct {
	productService services.ProductService
}

// NewProductController 是 ProductController 的构造函数
// 它负责获取 ProductService 的单例实例并注入
func NewProductController() *ProductController {
	// 获取产品服务的单例实例
	prodService, err := services.GetProductService()
	if err != nil {
		// 在实际生产环境中，这里应该使用日志记录错误，并可能采取更优雅的错误处理方式
		// 例如：返回一个全局的错误或使应用程序启动失败
		panic(fmt.Sprintf("初始化 ProductService 失败: %v", err))
	}
	return &ProductController{
		productService: prodService,
	}
}

// GetProducts 处理 GET /products 请求，获取所有产品
func (pc *ProductController) GetProducts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	products, err := pc.productService.GetAllProducts(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取产品列表", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"products": products})
}

// GetProductByID 处理 GET /products/:id 请求，根据ID获取单个产品
func (pc *ProductController) GetProductByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的产品ID"})
		return
	}

	product, err := pc.productService.GetProductByID(ctx, uint(id))
	if err != nil {
		// 根据错误类型返回不同的HTTP状态码
		if err.Error() == fmt.Sprintf("未找到产品 ID: %d", id) { // 检查是否是"未找到"错误
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("未找到产品 ID: %d", id)})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取产品", "details": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": product})
}

// CreateProduct 处理 POST /products 请求，创建新产品
func (pc *ProductController) CreateProduct(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var newProduct models.Product
	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	if err := pc.productService.CreateProduct(ctx, &newProduct); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建产品", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "产品已创建", "product": newProduct})
}

// UpdateProduct 处理 PUT /products/:id 请求，更新产品信息
func (pc *ProductController) UpdateProduct(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的产品ID"})
		return
	}

	var productToUpdate models.Product
	if err := c.ShouldBindJSON(&productToUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}
	productToUpdate.ID = uint(id) // 确保更新的是正确的ID

	if err := pc.productService.UpdateProduct(ctx, &productToUpdate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法更新产品", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "产品已更新", "product": productToUpdate})
}

// DeleteProduct 处理 DELETE /products/:id 请求，删除产品
func (pc *ProductController) DeleteProduct(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的产品ID"})
		return
	}

	if err := pc.productService.DeleteProduct(ctx, uint(id)); err != nil {
		if err.Error() == fmt.Sprintf("未找到产品 ID: %d 进行删除", id) { // 检查是否是"未找到"错误
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("未找到产品 ID: %d 进行删除", id)})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法删除产品", "details": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "产品已删除", "product_id": id})
}
