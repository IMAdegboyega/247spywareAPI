package repository

import (
	"blog-backend/internal/models"
	"time"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *models.Post) error {
	return r.db.Create(post).Error
}

func (r *PostRepository) FindByID(id uint) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("Author").Preload("Category").First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) FindBySlug(slug string) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("Author").Preload("Category").
		Where("slug = ?", slug).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) FindAll(page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	r.db.Model(&models.Post{}).Count(&total)

	offset := (page - 1) * perPage
	err := r.db.Preload("Author").Preload("Category").
		Order("created_at DESC").
		Offset(offset).Limit(perPage).
		Find(&posts).Error

	return posts, total, err
}

func (r *PostRepository) FindPublished(page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	r.db.Model(&models.Post{}).Where("status = ?", models.StatusPublished).Count(&total)

	offset := (page - 1) * perPage
	err := r.db.Preload("Author").Preload("Category").
		Where("status = ?", models.StatusPublished).
		Order("published_at DESC").
		Offset(offset).Limit(perPage).
		Find(&posts).Error

	return posts, total, err
}

func (r *PostRepository) FindLatestNews(limit int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Preload("Author").Preload("Category").
		Where("status = ? AND is_latest_news = ?", models.StatusPublished, true).
		Order("latest_news_at DESC").
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepository) FindByCategory(categoryID uint, page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	r.db.Model(&models.Post{}).
		Where("category_id = ? AND status = ?", categoryID, models.StatusPublished).
		Count(&total)

	offset := (page - 1) * perPage
	err := r.db.Preload("Author").Preload("Category").
		Where("category_id = ? AND status = ?", categoryID, models.StatusPublished).
		Order("published_at DESC").
		Offset(offset).Limit(perPage).
		Find(&posts).Error

	return posts, total, err
}

func (r *PostRepository) FindByAuthor(authorID uint, page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	r.db.Model(&models.Post{}).Where("author_id = ?", authorID).Count(&total)

	offset := (page - 1) * perPage
	err := r.db.Preload("Author").Preload("Category").
		Where("author_id = ?", authorID).
		Order("created_at DESC").
		Offset(offset).Limit(perPage).
		Find(&posts).Error

	return posts, total, err
}

func (r *PostRepository) FindScheduledDue() ([]models.Post, error) {
	var posts []models.Post
	now := time.Now()
	err := r.db.Where("status = ? AND scheduled_for <= ?", models.StatusScheduled, now).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepository) FindExpiredLatestNews(duration time.Duration) ([]models.Post, error) {
	var posts []models.Post
	cutoff := time.Now().Add(-duration)
	err := r.db.Where("is_latest_news = ? AND latest_news_at < ?", true, cutoff).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepository) Update(post *models.Post) error {
	return r.db.Save(post).Error
}

func (r *PostRepository) Delete(id uint) error {
	return r.db.Delete(&models.Post{}, id).Error
}

func (r *PostRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&models.Post{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *PostRepository) CountByStatus(status models.PostStatus) (int64, error) {
	var count int64
	err := r.db.Model(&models.Post{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *PostRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&models.Post{}).Count(&count).Error
	return count, err
}

func (r *PostRepository) SumViewCount() (int64, error) {
	var total int64
	err := r.db.Model(&models.Post{}).Select("COALESCE(SUM(view_count), 0)").Scan(&total).Error
	return total, err
}

func (r *PostRepository) Search(query string, page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	searchQuery := "%" + query + "%"

	r.db.Model(&models.Post{}).
		Where("status = ? AND (title LIKE ? OR content LIKE ?)", 
			models.StatusPublished, searchQuery, searchQuery).
		Count(&total)

	offset := (page - 1) * perPage
	err := r.db.Preload("Author").Preload("Category").
		Where("status = ? AND (title LIKE ? OR content LIKE ?)", 
			models.StatusPublished, searchQuery, searchQuery).
		Order("published_at DESC").
		Offset(offset).Limit(perPage).
		Find(&posts).Error

	return posts, total, err
}