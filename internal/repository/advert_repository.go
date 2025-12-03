package repository

import (
	"blog-backend/internal/models"
	"time"

	"gorm.io/gorm"
)

type AdvertRepository struct {
	db *gorm.DB
}

func NewAdvertRepository(db *gorm.DB) *AdvertRepository {
	return &AdvertRepository{db: db}
}

func (r *AdvertRepository) Create(advert *models.Advert) error {
	return r.db.Create(advert).Error
}

func (r *AdvertRepository) FindByID(id uint) (*models.Advert, error) {
	var advert models.Advert
	err := r.db.First(&advert, id).Error
	if err != nil {
		return nil, err
	}
	return &advert, nil
}

func (r *AdvertRepository) FindAll() ([]models.Advert, error) {
	var adverts []models.Advert
	err := r.db.Order("created_at DESC").Find(&adverts).Error
	return adverts, err
}

func (r *AdvertRepository) FindActive() ([]models.Advert, error) {
	var adverts []models.Advert
	now := time.Now()
	err := r.db.Where("is_active = ? AND (start_date IS NULL OR start_date <= ?) AND (end_date IS NULL OR end_date >= ?)", 
		true, now, now).
		Find(&adverts).Error
	return adverts, err
}

func (r *AdvertRepository) FindByPosition(position string) ([]models.Advert, error) {
	var adverts []models.Advert
	now := time.Now()
	err := r.db.Where("position = ? AND is_active = ? AND (start_date IS NULL OR start_date <= ?) AND (end_date IS NULL OR end_date >= ?)", 
		position, true, now, now).
		Find(&adverts).Error
	return adverts, err
}

func (r *AdvertRepository) Update(advert *models.Advert) error {
	return r.db.Save(advert).Error
}

func (r *AdvertRepository) Delete(id uint) error {
	return r.db.Delete(&models.Advert{}, id).Error
}

func (r *AdvertRepository) IncrementClickCount(id uint) error {
	return r.db.Model(&models.Advert{}).Where("id = ?", id).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}

func (r *AdvertRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&models.Advert{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}