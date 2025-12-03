package services

import (
	"errors"

	"blog-backend/internal/models"
	"blog-backend/internal/repository"
)

type AdvertService struct {
	advertRepo *repository.AdvertRepository
}

func NewAdvertService(advertRepo *repository.AdvertRepository) *AdvertService {
	return &AdvertService{advertRepo: advertRepo}
}

func (s *AdvertService) Create(req models.CreateAdvertRequest) (*models.Advert, error) {
	advert := &models.Advert{
		Title:     req.Title,
		ImageURL:  req.ImageURL,
		LinkURL:   req.LinkURL,
		Position:  req.Position,
		IsActive:  req.IsActive,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	if err := s.advertRepo.Create(advert); err != nil {
		return nil, err
	}

	return advert, nil
}

func (s *AdvertService) GetByID(id uint) (*models.Advert, error) {
	return s.advertRepo.FindByID(id)
}

func (s *AdvertService) GetAll() ([]models.Advert, error) {
	return s.advertRepo.FindAll()
}

func (s *AdvertService) GetActive() ([]models.Advert, error) {
	adverts, err := s.advertRepo.FindActive()
	if err != nil {
		return nil, err
	}

	// Increment view count for each advert returned
	for _, ad := range adverts {
		s.advertRepo.IncrementViewCount(ad.ID)
	}

	return adverts, nil
}

func (s *AdvertService) GetByPosition(position string) ([]models.Advert, error) {
	adverts, err := s.advertRepo.FindByPosition(position)
	if err != nil {
		return nil, err
	}

	// Increment view count
	for _, ad := range adverts {
		s.advertRepo.IncrementViewCount(ad.ID)
	}

	return adverts, nil
}

func (s *AdvertService) Update(id uint, req models.UpdateAdvertRequest) (*models.Advert, error) {
	advert, err := s.advertRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("advert not found")
	}

	if req.Title != "" {
		advert.Title = req.Title
	}
	if req.ImageURL != "" {
		advert.ImageURL = req.ImageURL
	}
	if req.LinkURL != "" {
		advert.LinkURL = req.LinkURL
	}
	if req.Position != "" {
		advert.Position = req.Position
	}
	if req.IsActive != nil {
		advert.IsActive = *req.IsActive
	}
	if req.StartDate != nil {
		advert.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		advert.EndDate = req.EndDate
	}

	if err := s.advertRepo.Update(advert); err != nil {
		return nil, err
	}

	return advert, nil
}

func (s *AdvertService) Delete(id uint) error {
	return s.advertRepo.Delete(id)
}

func (s *AdvertService) RecordClick(id uint) error {
	return s.advertRepo.IncrementClickCount(id)
}