package database

import (
	"log"

	"go-moon/internal/todo" // For todo.Todo model
	"go-moon/internal/user" // For user.User model
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger" // For GORM logger configuration
)

// ConnectDB initializes a connection to the PostgreSQL database and runs auto-migrations.
func ConnectDB(dsn string) (*gorm.DB, error) {
	// For more detailed logging, especially during development:
	// newLogger := logger.New(
	// 	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	// 	logger.Config{
	// 		SlowThreshold: time.Second, // Slow SQL threshold
	// 		LogLevel:      logger.Info, // Log level
	// 		Colorful:      true,        // Disable color
	// 	},
	// )

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Or logger.Info for more logs
		// Logger: newLogger, // Use custom logger
	})

	if err != nil {
		return nil, err
	}

	log.Println("Database connection successfully established.")

	// Auto-migrate schema
	// This will create tables, or add missing columns/indexes.
	// It will NOT delete unused columns, to protect your data.
	log.Println("Running database migrations...")
	err = db.AutoMigrate(
		&user.User{},
		&todo.Todo{},
		// Add other models here as they are created
	)
	if err != nil {
		log.Printf("Failed to auto-migrate database: %v\n", err)
		// Depending on the error, you might want to return it or handle it.
		// For critical migration errors, it might be best to halt.
		return nil, err
	}
	log.Println("Database migrations completed.")

	return db, nil
}

// Example DSN (Data Source Name) for local PostgreSQL:
// const DefaultDSN = "host=localhost user=youruser password=yourpassword dbname=yourdb port=5432 sslmode=disable TimeZone=Asia/Shanghai"
// You would typically get this from config.
// For this step, main.go will pass a hardcoded one or one from a simple config.
