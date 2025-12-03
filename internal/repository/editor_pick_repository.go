package repository

import (
	"blog-backend/internal/models"

	"gorm.io/gorm"
)

type EditorPickRepository struct {
	db *gorm.DB
}

func NewEditorPickRepository(db *gorm.DB) *EditorPickRepository {
	return &EditorPickRepository{db: db}
}

func (r *EditorPickRepository) Create(pick *models.EditorPick) error {
	return r.db.Create(pick).Error
}

func (r *EditorPickRepository) FindByID(id uint) (*models.EditorPick, error) {
	var pick models.EditorPick
	err := r.db.Preload("Post.Author").Preload("Post.Category").First(&pick, id).Error
	if err != nil {
		return nil, err
	}
	return &pick, nil
}

func (r *EditorPickRepository) FindByPostID(postID uint) (*models.EditorPick, error) {
	var pick models.EditorPick
	err := r.db.Where("post_id = ?", postID).First(&pick).Error
	if err != nil {
		return nil, err
	}
	return &pick, nil
}

func (r *EditorPickRepository) FindAll() ([]models.EditorPick, error) {
	var picks []models.EditorPick
	err := r.db.Preload("Post.Author").Preload("Post.Category").
		Order("picked_at DESC").
		Find(&picks).Error
	return picks, err
}

func (r *EditorPickRepository) Update(pick *models.EditorPick) error {
	return r.db.Save(pick).Error
}

func (r *EditorPickRepository) Delete(id uint) error {
	return r.db.Delete(&models.EditorPick{}, id).Error
}

func (r *EditorPickRepository) DeleteByPostID(postID uint) error {
	return r.db.Where("post_id = ?", postID).Delete(&models.EditorPick{}).Error
}

func (r *EditorPickRepository) UpdateOrder(picks []models.EditorPick) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, pick := range picks {
			if err := tx.Model(&models.EditorPick{}).
				Where("id = ?", pick.ID).
				Update("order", pick.Order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *EditorPickRepository) Exists(postID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.EditorPick{}).Where("post_id = ?", postID).Count(&count).Error
	return count > 0, err
}