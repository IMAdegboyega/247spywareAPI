package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/models"
	"blog-backend/internal/repository"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.Disconnect(ctx)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// --- Check if admin already exists ---------------------------------
	var existingAdmin models.User
	err = db.Database.Collection("users").
		FindOne(ctx, bson.M{"role": models.RoleAdmin}).
		Decode(&existingAdmin)
	if err == nil {
		fmt.Println("Admin user already exists:")
		fmt.Printf("  Email: %s\n", existingAdmin.Email)
		fmt.Printf("  Display Name: %s\n", existingAdmin.DisplayName)
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		log.Fatalf("Failed to query for admin: %v", err)
	} else {
		adminEmail := envOr("ADMIN_EMAIL", "admin@blog.com")
		adminPassword := envOr("ADMIN_PASSWORD", "admin123")
		adminName := envOr("ADMIN_NAME", "Admin")

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		userRepo := repository.NewUserRepository(db)
		admin := &models.User{
			Email:       adminEmail,
			Password:    string(hashedPassword),
			DisplayName: adminName,
			Role:        models.RoleAdmin,
			IsActive:    true,
		}
		if err := userRepo.Create(admin); err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}

		fmt.Println("Admin user created:")
		fmt.Printf("  Email: %s\n", adminEmail)
		fmt.Printf("  Password: %s\n", adminPassword)
		fmt.Println("⚠️  Please change the admin password after first login!")
	}

	// --- Seed default categories --------------------------------------
	categoryRepo := repository.NewCategoryRepository(db)
	defaults := []struct {
		Name string
		Slug string
	}{
		{"Technology", "technology"},
		{"Business", "business"},
		{"Lifestyle", "lifestyle"},
		{"Entertainment", "entertainment"},
		{"Sports", "sports"},
	}

	for _, c := range defaults {
		if _, err := categoryRepo.FindBySlug(c.Slug); err == nil {
			// already exists
			continue
		}
		cat := &models.Category{Name: c.Name, Slug: c.Slug}
		if err := categoryRepo.Create(cat); err != nil {
			log.Printf("Failed to create category %s: %v", c.Name, err)
			continue
		}
		fmt.Printf("Created category: %s\n", c.Name)
	}

	fmt.Println("\nSeeding complete!")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
