package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"blog-backend/internal/models"
	"blog-backend/internal/repository"
)

type SubscriberService struct {
	subscriberRepo *repository.SubscriberRepository
}

func NewSubscriberService(subscriberRepo *repository.SubscriberRepository) *SubscriberService {
	return &SubscriberService{subscriberRepo: subscriberRepo}
}

func (s *SubscriberService) Subscribe(email string) (*models.Subscriber, error) {
	// Check if already subscribed
	existing, _ := s.subscriberRepo.FindByEmail(email)
	if existing != nil {
		if existing.IsActive {
			return nil, errors.New("email already subscribed")
		}
		// Reactivate subscription
		existing.IsActive = true
		existing.SubscribedAt = time.Now()
		if err := s.subscriberRepo.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Generate unsubscribe token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	subscriber := &models.Subscriber{
		Email:            email,
		IsActive:         true,
		UnsubscribeToken: token,
		SubscribedAt:     time.Now(),
	}

	if err := s.subscriberRepo.Create(subscriber); err != nil {
		return nil, err
	}

	return subscriber, nil
}

func (s *SubscriberService) Unsubscribe(token string) error {
	subscriber, err := s.subscriberRepo.FindByToken(token)
	if err != nil {
		return errors.New("invalid unsubscribe token")
	}

	subscriber.IsActive = false
	return s.subscriberRepo.Update(subscriber)
}

func (s *SubscriberService) GetAll() ([]models.Subscriber, error) {
	return s.subscriberRepo.FindAll()
}

func (s *SubscriberService) Delete(id uint) error {
	return s.subscriberRepo.Delete(id)
}

func (s *SubscriberService) CountActive() (int64, error) {
	return s.subscriberRepo.CountActive()
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}