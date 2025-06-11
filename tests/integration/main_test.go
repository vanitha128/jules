package integration

import (
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"go-moon/internal/auth"
	"go-moon/internal/cache"
	"go-moon/internal/database"
	"go-moon/internal/middleware"
	"go-moon/internal/todo"
	"go-moon/internal/user"
	"go-moon/pkg/config"
	"go-moon/pkg/utils"
	"gorm.io/gorm"
)

var testDB *gorm.DB
var testCache cache.Cache // Define as the interface
var baseRouter *gin.Engine // A base router configured once

// TestMain sets up the test environment before running tests and tears it down afterward.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// 1. Load Config
	// Adjust path to your actual .env file or rely on environment variables for CI
	// It's often better to have a specific .env.test file or use environment variables.
	// For simplicity, this example tries to load from a path relative to where `go test` is run.
	// If running tests from the `tests/integration` directory, `../../.env` might be correct.
	// If running from project root, `.env` would be correct.
	// Defaulting to try loading from project root.
	cfg, err := config.LoadConfig(".env") // Or specific test config file/vars
	if err != nil {
		log.Fatalf("Failed to load test config: %v", err)
	}
	// Override DSN or Redis for test-specific databases if necessary
	// e.g., cfg.PostgresDSN = "host=localhost user=testuser password=testpass dbname=go_moon_test_db port=5432 sslmode=disable TimeZone=UTC"
	// e.g., cfg.RedisDB = 1 // Use a different Redis DB for tests

	// 2. Connect to DB
	testDB, err = database.ConnectDB(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to test DB: %v", err)
	}
	sqlDB, _ := testDB.DB()
	defer sqlDB.Close()

	// 3. Connect to Cache
	testCache, err = cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to test Redis: %v", err)
	}
	if c, ok := testCache.(interface{ Close() error }); ok {
		defer c.Close()
	}

	// 4. Setup Base Router (with all real dependencies)
	// This router is configured once and can be used by test cases.
	// Test cases should ensure data isolation (e.g., by clearing DB tables).
	jwtUtil := utils.NewJWTUtil(cfg.JWTSecret)
	authMw := middleware.AuthMiddleware(testCache, jwtUtil)

	userRepo := user.NewPostgresUserRepository(testDB)
	todoRepo := todo.NewPostgresTodoRepository(testDB)

	// Pass userRepo to NewAuthService as per earlier modification
	authSvc := auth.NewAuthService(testCache, jwtUtil, userRepo)
	userSvc := user.NewUserService(userRepo, jwtUtil)
	todoSvc := todo.NewTodoService(todoRepo)

	authHandler := auth.NewAuthHandler(authSvc)
	userHandler := user.NewUserHandler(userSvc)
	todoHandler := todo.NewTodoHandler(todoSvc)

	baseRouter = gin.Default() // Or gin.New() if you want minimal

	// Public routes
	publicUserRoutes := baseRouter.Group("/users")
	{
		publicUserRoutes.POST("/register", userHandler.RegisterUser)
		publicUserRoutes.POST("/login", userHandler.Login)
	}
	publicAuthRoutes := baseRouter.Group("/auth")
	{
		publicAuthRoutes.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected routes
	protectedUserRoutes := baseRouter.Group("/users")
	protectedUserRoutes.Use(authMw)
	{
		protectedUserRoutes.POST("/password", userHandler.ChangePassword)
		protectedUserRoutes.PUT("/me", userHandler.UpdateProfile)
		protectedUserRoutes.GET("/me/profile", userHandler.GetProfile) // Added GetProfile route
	}
	protectedAuthRoutes := baseRouter.Group("/auth")
	protectedAuthRoutes.Use(authMw)
	{
		protectedAuthRoutes.POST("/logout", authHandler.Logout)
	}
	todoProtectedRoutes := baseRouter.Group("/todos")
	todoProtectedRoutes.Use(authMw)
	{
		todoProtectedRoutes.POST("", todoHandler.CreateTodo)
		todoProtectedRoutes.GET("", todoHandler.ListTodos)
		todoProtectedRoutes.GET("/:todoID", todoHandler.GetTodo)
		todoProtectedRoutes.PUT("/:todoID", todoHandler.UpdateTodo)
		todoProtectedRoutes.DELETE("/:todoID", todoHandler.DeleteTodo)
	}

	// Run tests
	exitVal := m.Run()
	os.Exit(exitVal)
}

// clearDatabaseTables clears all data from user and todo tables.
// This should be called before each integration test or test suite that needs a clean DB.
func clearDatabaseTables() {
	// Order matters due to foreign keys if any are strictly enforced at DB level during delete
	// For GORM's default behavior, direct delete might be fine.
	// Using Unscoped() for Delete to ensure soft-deleted records are also cleared if GORM soft delete is ever enabled.
	if err := testDB.Exec("TRUNCATE TABLE todos CASCADE").Error; err != nil {
		log.Fatalf("Failed to truncate todos table: %v", err)
	}
	if err := testDB.Exec("TRUNCATE TABLE users CASCADE").Error; err != nil {
		log.Fatalf("Failed to truncate users table: %v", err)
	}
	// For a more robust cleanup, especially with many tables or complex relations:
	// 1. Disable foreign key checks
	// 2. Truncate all tables
	// 3. Re-enable foreign key checks
	// Example for PostgreSQL:
	// testDB.Exec("SET session_replication_role = 'replica';")
	// testDB.Exec("TRUNCATE TABLE users, todos, other_tables RESTART IDENTITY CASCADE;") // RESTART IDENTITY resets auto-increment counters
	// testDB.Exec("SET session_replication_role = 'origin';")
	log.Println("Database tables cleared.")
}

// clearRedisCache (if needed and safe for your test Redis instance)
func clearRedisCache() {
	// The `FLUSHDB` command can be used if this Redis instance/DB is dedicated to testing.
	// Be very careful with this command on a shared Redis.
	// Example (requires direct redis client if not exposed by Cache interface):
	// if rc, ok := testCache.(*cache.redisCache); ok { // Type assertion to access underlying client
	//    client := rc.GetClient() // Assume GetClient() exists or testCache is the client itself
	//    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//    defer cancel()
	//    if err := client.FlushDB(ctx).Err(); err != nil {
	//        log.Fatalf("Failed to flush Redis test DB: %v", err)
	//    }
	//    log.Println("Redis test DB flushed.")
	// }
	// For now, manual clearing or relying on test isolation by key prefixing is safer if FlushDB is risky.
	// The `cache` interface does not expose `FlushDB`.
	// Individual tests are responsible for cleaning up keys they create if global flush is not used.
	log.Println("Redis cache clear function called (currently a no-op, implement if safe).")
}

// setupTestRouter is a helper that can be called by individual tests if they need a fresh router
// instance or if TestMain's baseRouter isn't sufficient for some reason (e.g. tests that modify router state).
// For most cases, using the baseRouter from TestMain and clearing data should be enough.
// This function is similar to the router setup in TestMain.
func setupTestRouter() *gin.Engine {
    // This function would re-run the router setup logic from TestMain.
    // For this project, we will rely on the baseRouter and data clearing.
    // If specific tests need a radically different router setup, they can define it locally.
    // For now, just return the global baseRouter, assuming data clearing is handled.
    return baseRouter
}
