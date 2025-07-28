// internal/controllers/user_controller.go
package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"

	"github.com/MGJYW/QSLCardManageServer/internal/models"
	"github.com/MGJYW/QSLCardManageServer/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	userService services.UserService
}

func NewUserController(userService services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (uc *UserController) CreateUser(c *gin.Context) {
	h := sha256.New()
	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if newUser.Email == nil && newUser.Phone == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法创建用户", "details": "邮箱与电话必须存在任意一个"})
		return
	}
	if newUser.Name == "" {
		newUser.Name = "猫猫"
	}
	newUUID, err := uuid.NewRandom()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建用户", "details": err.Error()})
		return
	}
	newUser.UserID = newUUID.String()

	log.Println(newUser.Password)
	h.Write([]byte(newUser.Password))
	newUser.Password = hex.EncodeToString(h.Sum(nil))
	if err := uc.userService.CreateUser(c.Request.Context(), &newUser); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "无法创建用户", "details": err.Error()})
		return
	}

	userJwt, err := uc.userService.GenerateToken(c.Request.Context(), &newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建用户", "details": err.Error()})
		return
	}
	newUser.Password = ""
	newUser.IdCardR = ""
	c.JSON(http.StatusCreated, gin.H{
		"message": "用户注册成功",
		"token":   userJwt,
		"user": gin.H{ // 只返回非敏感的用户信息
			"userId": newUser.UserID,
			"name":   newUser.Name,
			"email":  newUser.Email,
			"phone":  newUser.Phone,
			// 不要返回 HashedPassword 或 IdCardR
		},
	})
}

func (uc *UserController) LoginUser(c *gin.Context) {
	var user models.UserLogin
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user_res, err := uc.userService.GetUserByText(c.Request.Context(), user.Loginfo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "登录失败", "details": err.Error()})
		return
	}
	h := sha256.New()
	h.Write([]byte(user.Password))
	log.Println(user.Password)
	log.Println(hex.EncodeToString(h.Sum(nil)))
	log.Println(user_res.Password)
	if user_res.Password != hex.EncodeToString(h.Sum(nil)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "登录失败", "details": "密码不正确"})
		return
	}
	userJwt, err := uc.userService.GenerateToken(c.Request.Context(), user_res)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建令牌", "details": err.Error()})
		return
	}
	user_res.Password = ""
	user_res.IdCardR = ""
	c.JSON(http.StatusCreated, gin.H{
		"message": "用户注册成功",
		"token":   userJwt,
		"user": gin.H{ // 只返回非敏感的用户信息
			"userId": user_res.UserID,
			"name":   user_res.Name,
			"email":  user_res.Email,
			"phone":  user_res.Phone,
			// 不要返回 HashedPassword 或 IdCardR
		},
	})
}

func (uc *UserController) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	user, err := uc.userService.GetUserByText(c.Request.Context(), idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "无法获取用户", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "获取用户成功", "userinfo": user})
}

func (uc *UserController) Testtoken(c *gin.Context) {
	a, err := uc.userService.VerifyToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌验证失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "成功", "user": a})
}
