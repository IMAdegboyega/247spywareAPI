package main

import (
	"fmt"
	"log"
	"os"

	"blog-backend/internal/config"
	"blog-backend/internal/models"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Post{},
		&models.EditorPick{},
		&models.Subscriber{},
		&models.Advert{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Check if admin already exists
	var existingAdmin models.User
	result := db.Where("role = ?", models.RoleAdmin).First(&existingAdmin)
	if result.Error == nil {
		fmt.Println("Admin user already exists:")
		fmt.Printf("  Email: %s\n", existingAdmin.Email)
		fmt.Printf("  Display Name: %s\n", existingAdmin.DisplayName)
		return
	}

	// Create default admin user
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@blog.com"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
	}

	adminName := os.Getenv("ADMIN_NAME")
	if adminName == "" {
		adminName = "Admin"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	admin := models.User{
		Email:       adminEmail,
		Password:    string(hashedPassword),
		DisplayName: adminName,
		Role:        models.RoleAdmin,
		IsActive:    true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Println("Database seeded successfully!")
	fmt.Println("Admin user created:")
	fmt.Printf("  Email: %s\n", adminEmail)
	fmt.Printf("  Password: %s\n", adminPassword)
	fmt.Println("\n⚠️  Please change the admin password after first login!")

	// Create some default categories
	categories := []models.Category{
		{Name: "Technology", Slug: "technology"},
		{Name: "Business", Slug: "business"},
		{Name: "Lifestyle", Slug: "lifestyle"},
		{Name: "Entertainment", Slug: "entertainment"},
		{Name: "Sports", Slug: "sports"},
	}

	for _, cat := range categories {
		var existing models.Category
		if db.Where("slug = ?", cat.Slug).First(&existing).Error != nil {
			db.Create(&cat)
			fmt.Printf("Created category: %s\n", cat.Name)
		}
	}

	fmt.Println("\nSeeding complete!")
}