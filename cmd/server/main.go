package main

import (
	"context"
	"log"
	"os"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/handlers"
	"blog-backend/internal/middleware"
	"blog-backend/internal/repository"
	"blog-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.Disconnect(ctx)
	}()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	postRepo := repository.NewPostRepository(db)
	editorPickRepo := repository.NewEditorPickRepository(db)
	subscriberRepo := repository.NewSubscriberRepository(db)
	advertRepo := repository.NewAdvertRepository(db)

	// Shared email service — handles both notification (Gmail) and newsletter (Resend) senders
	emailService := services.NewEmailService()

	// Initialize services
	authService := services.NewAuthService(userRepo, postRepo, emailService, os.Getenv("JWT_SECRET"))
	categoryService := services.NewCategoryService(categoryRepo, postRepo)
	postService := services.NewPostService(postRepo, subscriberRepo, userRepo, categoryRepo, emailService)
	editorPickService := services.NewEditorPickService(editorPickRepo, postRepo)
	subscriberService := services.NewSubscriberService(subscriberRepo, emailService)
	advertService := services.NewAdvertService(advertRepo)
	uploadService := services.NewUploadService(os.Getenv("UPLOAD_PATH"))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	postHandler := handlers.NewPostHandler(postService, uploadService, categoryService)
	editorPickHandler := handlers.NewEditorPickHandler(editorPickService)
	subscriberHandler := handlers.NewSubscriberHandler(subscriberService)
	advertHandler := handlers.NewAdvertHandler(advertService)

	// Setup router
	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORS())

	// Static files for uploads
	router.Static("/uploads", os.Getenv("UPLOAD_PATH"))

	// Public routes
	public := router.Group("/api")
	public.Use(middleware.RateLimit(120, 20))
	{
		// Auth (stricter limit)
		public.POST("/auth/login", middleware.StrictRateLimit(10, 5), authHandler.Login)
		public.POST("/auth/forgot-password", middleware.StrictRateLimit(5, 3), authHandler.ForgotPassword)
		public.POST("/auth/reset-password", middleware.StrictRateLimit(10, 5), authHandler.ResetPassword)

		// Public blog routes
		public.GET("/posts", postHandler.GetPublishedPosts)
		public.GET("/posts/:slug", postHandler.GetPostBySlug)
		public.GET("/posts/latest", postHandler.GetLatestNews)
		public.GET("/posts/category/:slug", postHandler.GetPostsByCategory)
		public.GET("/categories", categoryHandler.GetAll)
		public.GET("/editor-picks", editorPickHandler.GetAll)
		public.GET("/adverts/active", advertHandler.GetActive)

		// Newsletter subscription (stricter limit)
		public.POST("/subscribe", middleware.StrictRateLimit(5, 3), subscriberHandler.Subscribe)
		public.GET("/unsubscribe", subscriberHandler.Unsubscribe)
	}

	// Protected routes (requires authentication)
	protected := router.Group("/api/admin")
	protected.Use(middleware.AuthMiddleware(os.Getenv("JWT_SECRET")))
	{
		// User management (admin only)
		protected.POST("/users", middleware.AdminOnly(), authHandler.CreateUser)
		protected.GET("/users", middleware.AdminOnly(), authHandler.GetAllUsers)
		protected.PUT("/users/:id", middleware.AdminOnly(), authHandler.UpdateUser)
		protected.DELETE("/users/:id", middleware.AdminOnly(), authHandler.DeleteUser)
		protected.PUT("/users/:id/toggle-active", middleware.AdminOnly(), authHandler.ToggleUserActive)

		// Category management (admin only)
		protected.POST("/categories", middleware.AdminOnly(), categoryHandler.Create)
		protected.PUT("/categories/:id", middleware.AdminOnly(), categoryHandler.Update)
		protected.DELETE("/categories/:id", middleware.AdminOnly(), categoryHandler.Delete)

		// Post management
		protected.GET("/posts", postHandler.GetAllPosts)
		protected.GET("/posts/:id", postHandler.GetPostByID)
		protected.POST("/posts", postHandler.Create)
		protected.PUT("/posts/:id", postHandler.Update)
		protected.DELETE("/posts/:id", middleware.AdminOnly(), postHandler.Delete)
		protected.PUT("/posts/:id/status", postHandler.UpdateStatus)
		protected.PUT("/posts/:id/auto-save", postHandler.AutoSave)
		protected.POST("/posts/:id/upload-image", postHandler.UploadImage)

		// Editor's picks (admin only)
		protected.POST("/editor-picks", middleware.AdminOnly(), editorPickHandler.Add)
		protected.DELETE("/editor-picks/:id", middleware.AdminOnly(), editorPickHandler.Remove)
		protected.PUT("/editor-picks/reorder", middleware.AdminOnly(), editorPickHandler.Reorder)

		// Subscribers (admin only)
		protected.GET("/subscribers", middleware.AdminOnly(), subscriberHandler.GetAll)
		protected.DELETE("/subscribers/:id", middleware.AdminOnly(), subscriberHandler.Delete)

		// Adverts (admin only)
		protected.GET("/adverts", middleware.AdminOnly(), advertHandler.GetAll)
		protected.POST("/adverts", middleware.AdminOnly(), advertHandler.Create)
		protected.PUT("/adverts/:id", middleware.AdminOnly(), advertHandler.Update)
		protected.DELETE("/adverts/:id", middleware.AdminOnly(), advertHandler.Delete)

		// Dashboard stats
		protected.GET("/stats", middleware.AdminOnly(), postHandler.GetStats)
	}

	// Start scheduled tasks (check for scheduled posts, expire latest news)
	go services.StartScheduler(postService)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
