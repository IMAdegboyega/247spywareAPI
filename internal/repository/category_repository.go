package repository

import (
	"errors"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CategoryRepository struct {
	db   *config.MongoDB
	coll *mongo.Collection
}

func NewCategoryRepository(db *config.MongoDB) *CategoryRepository {
	return &CategoryRepository{db: db, coll: db.Database.Collection("categories")}
}

func (r *CategoryRepository) Create(category *models.Category) error {
	ctx, cancel := reqCtx()
	defer cancel()

	id, err := r.db.NextSequence(ctx, "categories")
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	category.ID = id
	category.CreatedAt = now
	category.UpdatedAt = now

	_, err = r.coll.InsertOne(ctx, category)
	return err
}

func (r *CategoryRepository) FindByID(id uint) (*models.Category, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var cat models.Category
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&cat)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepository) FindBySlug(slug string) (*models.Category, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var cat models.Category
	err := r.coll.FindOne(ctx, bson.M{"slug": slug}).Decode(&cat)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepository) FindAll() ([]models.Category, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cur, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var cats []models.Category
	if err := cur.All(ctx, &cats); err != nil {
		return nil, err
	}
	return cats, nil
}

func (r *CategoryRepository) Update(category *models.Category) error {
	ctx, cancel := reqCtx()
	defer cancel()

	category.UpdatedAt = time.Now().UTC()
	_, err := r.coll.ReplaceOne(ctx, bson.M{"_id": category.ID}, category)
	return err
}

func (r *CategoryRepository) Delete(id uint) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
