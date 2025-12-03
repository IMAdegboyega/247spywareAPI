package repository

import (
	"blog-backend/internal/models"

	"gorm.io/gorm"
)

type SubscriberRepository struct {
	db *gorm.DB
}

func NewSubscriberRepository(db *gorm.DB) *SubscriberRepository {
	return &SubscriberRepository{db: db}
}

func (r *SubscriberRepository) Create(subscriber *models.Subscriber) error {
	return r.db.Create(subscriber).Error
}

func (r *SubscriberRepository) FindByEmail(email string) (*models.Subscriber, error) {
	var subscriber models.Subscriber
	err := r.db.Where("email = ?", email).First(&subscriber).Error
	if err != nil {
		return nil, err
	}
	return &subscriber, nil
}

func (r *SubscriberRepository) FindByToken(token string) (*models.Subscriber, error) {
	var subscriber models.Subscriber
	err := r.db.Where("unsubscribe_token = ?", token).First(&subscriber).Error
	if err != nil {
		return nil, err
	}
	return &subscriber, nil
}

func (r *SubscriberRepository) FindByID(id uint) (*models.Subscriber, error) {
	var subscriber models.Subscriber
	err := r.db.First(&subscriber, id).Error
	if err != nil {
		return nil, err
	}
	return &subscriber, nil
}

func (r *SubscriberRepository) FindAllActive() ([]models.Subscriber, error) {
	var subscribers []models.Subscriber
	err := r.db.Where("is_active = ?", true).Order("subscribed_at DESC").Find(&subscribers).Error
	return subscribers, err
}

func (r *SubscriberRepository) FindAll() ([]models.Subscriber, error) {
	var subscribers []models.Subscriber
	err := r.db.Order("subscribed_at DESC").Find(&subscribers).Error
	return subscribers, err
}

func (r *SubscriberRepository) Update(subscriber *models.Subscriber) error {
	return r.db.Save(subscriber).Error
}

func (r *SubscriberRepository) Delete(id uint) error {
	return r.db.Delete(&models.Subscriber{}, id).Error
}

func (r *SubscriberRepository) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&models.Subscriber{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}