package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"online-lib/internal/config"
	"online-lib/internal/database"
	"online-lib/internal/handlers"
	"online-lib/internal/middleware"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("не удалось подключиться к базе данных: %v", err)
	}

	auth := &handlers.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	books := &handlers.BookHandler{DB: db}

	r := gin.Default()
	r.Use(middleware.CORS(cfg.AllowedOrigin))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/auth/register", auth.Register)
	r.POST("/auth/login", auth.Login)
	r.GET("/auth/me", middleware.Auth(cfg.JWTSecret), auth.Me)

	r.GET("/books", books.List)
	r.GET("/books/:id", books.Get)
	r.POST("/books", middleware.Auth(cfg.JWTSecret), books.Create)
	r.PATCH("/books/:id", middleware.Auth(cfg.JWTSecret), books.Update)
	r.DELETE("/books/:id", middleware.Auth(cfg.JWTSecret), books.Delete)

	r.StaticFile("/swagger", "./docs/index.html")
	r.StaticFile("/swagger/openapi.json", "./docs/openapi.json")

	log.Printf("API доступен на http://localhost:%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}
}
