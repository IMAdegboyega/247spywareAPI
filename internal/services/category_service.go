package services

import (
	"errors"
	"regexp"
	"strings"

	"blog-backend/internal/models"
	"blog-backend/internal/repository"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

func (s *CategoryService) Create(req models.CreateCategoryRequest) (*models.Category, error) {
	slug := generateSlug(req.Name)

	// Check if slug already exists
	existing, _ := s.categoryRepo.FindBySlug(slug)
	if existing != nil {
		return nil, errors.New("category with this name already exists")
	}

	category := &models.Category{
		Name: req.Name,
		Slug: slug,
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) GetByID(id uint) (*models.Category, error) {
	return s.categoryRepo.FindByID(id)
}

func (s *CategoryService) GetBySlug(slug string) (*models.Category, error) {
	return s.categoryRepo.FindBySlug(slug)
}

func (s *CategoryService) GetAll() ([]models.Category, error) {
	return s.categoryRepo.FindAll()
}

func (s *CategoryService) Update(id uint, req models.UpdateCategoryRequest) (*models.Category, error) {
	category, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("category not found")
	}

	newSlug := generateSlug(req.Name)

	// Check if new slug is taken by another category
	existing, _ := s.categoryRepo.FindBySlug(newSlug)
	if existing != nil && existing.ID != id {
		return nil, errors.New("category with this name already exists")
	}

	category.Name = req.Name
	category.Slug = newSlug

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) Delete(id uint) error {
	return s.categoryRepo.Delete(id)
}

// generateSlug creates a URL-friendly slug from a string
func generateSlug(s string) string {
	// Convert to lowercase
	slug := strings.ToLower(s)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}