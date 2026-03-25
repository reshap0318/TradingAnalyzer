package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/reshap/trading-bot/cmd/migration/seeders"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get database connection string from environment
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("Database configuration is incomplete. Please set DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, and DB_NAME")
	}

	// Build MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Parse command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "up":
		runMigration(dsn, "up")
	case "down":
		runMigration(dsn, "down")
	case "seed":
		runSeed(dsn)
	case "refresh":
		runMigration(dsn, "down")
		runMigration(dsn, "up")
		runSeed(dsn)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: go run cmd/migration/main.go <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up       Apply all migrations (create/update tables)")
	fmt.Println("  down     Drop all migrated tables")
	fmt.Println("  seed     Insert default data")
	fmt.Println("  refresh  Drop all tables, recreate, and seed data")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migration/main.go up")
	fmt.Println("  go run cmd/migration/main.go down")
	fmt.Println("  go run cmd/migration/main.go seed")
	fmt.Println("  go run cmd/migration/main.go refresh")
}

func runMigration(dsn string, command string) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	switch command {
	case "up":
		fmt.Println("Running migrations...")

		// AutoMigrate all models
		err := db.AutoMigrate(
			// Master data
			&models.Threshold{},
			&models.Timeframe{},
			&models.Config{},
			&models.Indicators{},
			// Strategy
			&models.Strategy{},
			&models.StrategyTimeframe{},
			&models.StrategyIndicator{},
			&models.StrategyMoneyMgmt{},
			&models.StrategySymbol{},
			// Transactional
			&models.Watchlist{},
			&models.Trade{},
			&models.TradeEntry{},
			&models.Backtest{},
			&models.BacktestTrade{},
			// Signal
			&models.Signal{},
		)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

		fmt.Println("✓ Migration completed successfully!")

	case "down":
		fmt.Println("Rolling back migrations...")

		// Drop tables in correct order (foreign key constraints)
		err := db.Migrator().DropTable(
			// Transactional (has foreign keys)
			&models.BacktestTrade{},
			&models.TradeEntry{},
			&models.Trade{},
			&models.Watchlist{},
			// Signal
			&models.Signal{},
			// Strategy (has foreign keys)
			&models.StrategyMoneyMgmt{},
			&models.StrategyIndicator{},
			&models.StrategyTimeframe{},
			&models.StrategySymbol{},
			&models.Strategy{},
			// Master data
			&models.Indicators{},
			&models.Config{},
			&models.Timeframe{},
			&models.Threshold{},
			// Backtest
			&models.Backtest{},
		)
		if err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}

		fmt.Println("✓ Rollback completed successfully!")
	}
}

func runSeed(dsn string) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("\n🌱 Seeding default data...\n")

	// Seed in correct order (respecting foreign key constraints)
	seeders.SeedThreshold(db)
	seeders.SeedTimeframes(db)
	seeders.SeedConfig(db)
	seeders.SeedIndicators(db)
	seeders.SeedStrategies(db)
	seeders.SeedWatchlist(db)

	fmt.Println("\n✅ Seeding completed!")
}
