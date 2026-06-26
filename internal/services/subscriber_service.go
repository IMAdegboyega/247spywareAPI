package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"blog-backend/internal/models"
	"blog-backend/internal/repository"
)

type SubscriberService struct {
	subscriberRepo *repository.SubscriberRepository
	email          *EmailService
}

func NewSubscriberService(subscriberRepo *repository.SubscriberRepository, email *EmailService) *SubscriberService {
	return &SubscriberService{
		subscriberRepo: subscriberRepo,
		email:          email,
	}
}

func (s *SubscriberService) Subscribe(email string) (*models.Subscriber, error) {
	// Check if already subscribed
	existing, _ := s.subscriberRepo.FindByEmail(email)
	if existing != nil {
		if existing.IsActive {
			return nil, errors.New("email already subscribed")
		}
		// Reactivate subscription — send a "welcome back" mail so they get
		// the same confirmation experience as a fresh signup.
		existing.IsActive = true
		existing.SubscribedAt = time.Now()
		if err := s.subscriberRepo.Update(existing); err != nil {
			return nil, err
		}
		go s.sendWelcome(existing)
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

	go s.sendWelcome(subscriber)

	return subscriber, nil
}

// sendWelcome fires the welcome email off the request goroutine so a slow SMTP
// hop never makes the subscribe response hang.
func (s *SubscriberService) sendWelcome(sub *models.Subscriber) {
	if s.email == nil || !s.email.IsNewsletterEnabled() {
		return
	}
	if err := s.email.SendWelcomeEmail(sub.Email, sub.UnsubscribeToken); err != nil {
		log.Printf("subscribe: welcome email to %s failed: %v", sub.Email, err)
	}
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
