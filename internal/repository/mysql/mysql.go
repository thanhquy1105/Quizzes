package mysql

import (
	"database/sql"
	"fmt"
	"os"

	"btaskee-quiz/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(cfg *config.MySQLConfig) (*gorm.DB, error) {
	// First, connect to MySQL without a database name to create it if it doesn't exist
	dsnNoDB := fmt.Sprintf("%s:%s@tcp(%s)/?parseTime=true",
		cfg.User, cfg.Password, cfg.Addr)
	dbRaw, err := sql.Open("mysql", dsnNoDB)
	if err != nil {
		return nil, fmt.Errorf("failed to open raw mysql connection: %w", err)
	}
	defer dbRaw.Close()

	_, err = dbRaw.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.DBName))
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Now connect with GORM to the specific database
	// Enable multiStatements=true to execute the entire seed.sql file at once
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&multiStatements=true",
		cfg.User, cfg.Password, cfg.Addr, cfg.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mysql with gorm: %w", err)
	}

	// Only migrate and seed if the database is empty (checked by existence of quizzes table)
	if !db.Migrator().HasTable("quizzes") {
		fmt.Println("Database is empty, executing seed/seed.sql...")
		sqlFile, err := os.ReadFile("seed/seed.sql")
		if err != nil {
			return nil, fmt.Errorf("failed to read seed/seed.sql: %w", err)
		}

		// Execute the entire SQL file
		// Note: Some drivers might need splitting by ';', but with multiStatements=true it should work
		err = db.Exec(string(sqlFile)).Error
		if err != nil {
			// If executing the whole file fails, try splitting if needed, 
			// but for now we trust multiStatements=true
			return nil, fmt.Errorf("failed to execute seed/seed.sql: %w", err)
		}
		fmt.Println("Database initialized and seeded successfully.")
	}

	return db, nil
}
