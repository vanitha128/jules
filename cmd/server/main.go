package main

import (
	"log"
	"net/http" // For http.StatusOK in placeholder if needed
	"strings"  // For strings.Replace in placeholder logout
	"time"     // For time.Minute in placeholder logout

	"github.com/your-username/go-moon/internal/auth"       // Auth package
	"github.com/your-username/go-moon/internal/cache"      // Cache package
	"github.com/your-username/go-moon/internal/database"   // Actual DB package
	"github.com/your-username/go-moon/internal/middleware" // Middleware package
	"github.com/your-username/go-moon/internal/todo"       // Todo package
	"github.com/your-username/go-moon/internal/user"       // User package
	"github.com/your-username/go-moon/pkg/config"     // Config package
	"github.com/your-username/go-moon/pkg/utils"      // JWT Utility

	"github.com/gin-gonic/gin"
	// "gorm.io/gorm" // Not needed directly here if db package handles it
)

func main() {
	// --- Load Configuration ---
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// --- Initialize Database ---
	db, err := database.ConnectDB(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	sqlDB, err := db.DB() // Get underlying sql.DB for Close
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()
	log.Println("Database connection established and migrations run.")

	// --- Initialize Cache ---
	appCache, err := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	// Check if appCache implements a Close method (it does via redisCache struct)
	if c, ok := appCache.(interface{ Close() error }); ok {
		defer c.Close()
		log.Println("Redis cache connection established.")
	} else {
		log.Println("Redis cache connection established (Close method not found or not applicable).")
	}


	// --- Initialize Utilities ---
	jwtUtil := utils.NewJWTUtil(cfg.JWTSecret)

	// --- Initialize Repositories ---
	userRepo := user.NewPostgresUserRepository(db)
	todoRepo := todo.NewPostgresTodoRepository(db)

	// --- Initialize Services ---
	// Note: The Login method in UserService currently uses utils.GenerateAccessToken/RefreshToken directly.
	// This should be updated to use the jwtUtil instance or delegate login fully to AuthService.
	// For now, we proceed with this structure, but it's a point of refactoring.
	authService := auth.NewAuthService(appCache, jwtUtil)
	userService := user.NewUserService(userRepo, jwtUtil) // Pass jwtUtil
	todoService := todo.NewTodoService(todoRepo)

	// --- Initialize Handlers ---
	userHandler := user.NewUserHandler(userService) // UserHandler's Login method needs to be aware of JWTUtil or AuthService
	authHandler := auth.NewAuthHandler(authService)
	todoHandler := todo.NewTodoHandler(todoService)

	// --- Setup Router ---
	router := gin.Default()
	// Initialize AuthMiddleware with dependencies
	authMw := middleware.AuthMiddleware(appCache, jwtUtil)

	// --- Public Routes ---
	// Group for user registration and login
	publicUserRoutes := router.Group("/users")
	{
		publicUserRoutes.POST("/register", userHandler.RegisterUser)
		publicUserRoutes.POST("/login", userHandler.Login) // This handler internally calls user service's Login
	}

	publicAuthRoutes := router.Group("/auth")
	{
		publicAuthRoutes.POST("/refresh", authHandler.RefreshToken) // Refresh token endpoint
	}

	// --- Protected Routes ---
	// Group for authenticated user actions (password change, profile update)
	protectedUserRoutes := router.Group("/users")
	protectedUserRoutes.Use(authMw)
	{
		protectedUserRoutes.POST("/password", userHandler.ChangePassword)
		protectedUserRoutes.PUT("/me", userHandler.UpdateProfile)
	}

	// Group for authenticated auth actions (logout)
	protectedAuthRoutes := router.Group("/auth")
	protectedAuthRoutes.Use(authMw)
	{
		protectedAuthRoutes.POST("/logout", authHandler.Logout)
	}

	// Group for TODO items (all protected)
	todoRoutes := router.Group("/todos")
	todoRoutes.Use(authMw)
	{
		todoRoutes.POST("/", todoHandler.CreateTodo)
		todoRoutes.PUT("/:todoID", todoHandler.UpdateTodo)
		todoRoutes.GET("/", todoHandler.ListTodos)
		todoRoutes.GET("/:todoID", todoHandler.GetTodo)
		todoRoutes.DELETE("/:todoID", todoHandler.DeleteTodo)
	}

	log.Printf("Server starting on port %s...", cfg.ServerPort)
	if err := router.Run(cfg.ServerPort); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
