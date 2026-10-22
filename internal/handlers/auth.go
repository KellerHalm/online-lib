package handlers

import (
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"online-lib/internal/models"
)

type AuthHandler struct {
	DB        *gorm.DB
	JWTSecret string
}

type registerInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var in registerInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидное тело запроса, ожидается JSON"})
		return
	}
	email := strings.TrimSpace(strings.ToLower(in.Email))
	var errs []string
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		errs = append(errs, "email обязателен и должен быть корректным")
	}
	if len(in.Password) < 6 {
		errs = append(errs, "пароль не может быть короче 6 символов")
	}
	if len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": strings.Join(errs, "; ")})
		return
	}
	var existing models.User
	if err := h.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "пользователь с таким email уже зарегистрирован"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать пользователя"})
		return
	}
	user := models.User{Email: email, Password: string(hash)}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "пользователь с таким email уже зарегистрирован"})
		return
	}
	c.JSON(http.StatusCreated, models.User{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var in registerInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидное тело запроса, ожидается JSON"})
		return
	}
	email := strings.TrimSpace(strings.ToLower(in.Email))
	var user models.User
	if err := h.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный email или пароль"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный email или пароль"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось выдать токен"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"accessToken": tokenStr})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetUint("userID")
	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "недействительный или просроченный токен"})
		return
	}
	c.JSON(http.StatusOK, models.User{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt})
}
