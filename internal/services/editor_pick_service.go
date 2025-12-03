package services

import (
	"errors"
	"time"

	"blog-backend/internal/models"
	"blog-backend/internal/repository"
)

type EditorPickService struct {
	editorPickRepo *repository.EditorPickRepository
	postRepo       *repository.PostRepository
}

func NewEditorPickService(editorPickRepo *repository.EditorPickRepository, postRepo *repository.PostRepository) *EditorPickService {
	return &EditorPickService{
		editorPickRepo: editorPickRepo,
		postRepo:       postRepo,
	}
}

func (s *EditorPickService) Add(postID uint) (*models.EditorPick, error) {
	// Check if post exists
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// Only published posts can be editor's picks
	if post.Status != models.StatusPublished {
		return nil, errors.New("only published posts can be editor's picks")
	}

	// Check if already an editor's pick
	exists, _ := s.editorPickRepo.Exists(postID)
	if exists {
		return nil, errors.New("post is already an editor's pick")
	}

	pick := &models.EditorPick{
		PostID:   postID,
		PickedAt: time.Now(),
	}

	if err := s.editorPickRepo.Create(pick); err != nil {
		return nil, err
	}

	return s.editorPickRepo.FindByID(pick.ID)
}

func (s *EditorPickService) GetAll() ([]models.EditorPick, error) {
	return s.editorPickRepo.FindAll()
}

func (s *EditorPickService) Remove(id uint) error {
	return s.editorPickRepo.Delete(id)
}

func (s *EditorPickService) RemoveByPostID(postID uint) error {
	return s.editorPickRepo.DeleteByPostID(postID)
}

func (s *EditorPickService) Reorder(req models.ReorderEditorPicksRequest) error {
	var picks []models.EditorPick
	for _, p := range req.Picks {
		picks = append(picks, models.EditorPick{
			ID:    p.ID,
			Order: p.Order,
		})
	}
	return s.editorPickRepo.UpdateOrder(picks)
}