package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/acmecorp/platform-api/internal/auth"
	"github.com/acmecorp/platform-api/internal/database"
	"github.com/acmecorp/platform-api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	db, err := database.Connect(viper.GetString("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	jwtService := auth.NewJWTService(viper.GetString("JWT_SECRET"))

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "db": db.Health()})
	})

	api := r.Group("/api/v1")
	api.Use(middleware.AuthRequired(jwtService))
	{
		api.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"users": []string{}})
		})
		api.POST("/users", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"id": "usr-001"})
		})
	}

	srv := &http.Server{
		Addr:    ":" + viper.GetString("PORT"),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
