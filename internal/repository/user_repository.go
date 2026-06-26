package repository

import (
	"context"
	"errors"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrNotFound is returned when a document doesn't exist. Services check for
// err != nil already, so this is mostly informational.
var ErrNotFound = errors.New("not found")

type UserRepository struct {
	db   *config.MongoDB
	coll *mongo.Collection
}

func NewUserRepository(db *config.MongoDB) *UserRepository {
	return &UserRepository{db: db, coll: db.Database.Collection("users")}
}

func reqCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func (r *UserRepository) Create(user *models.User) error {
	ctx, cancel := reqCtx()
	defer cancel()

	id, err := r.db.NextSequence(ctx, "users")
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now
	if user.Role == "" {
		user.Role = models.RoleAuthor
	}

	_, err = r.coll.InsertOne(ctx, user)
	return err
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var user models.User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var user models.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll() ([]models.User, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var users []models.User
	if err := cur.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// Update writes back the full user document.
func (r *UserRepository) Update(user *models.User) error {
	ctx, cancel := reqCtx()
	defer cancel()

	user.UpdatedAt = time.Now().UTC()
	_, err := r.coll.ReplaceOne(ctx, bson.M{"_id": user.ID}, user)
	return err
}

func (r *UserRepository) Delete(id uint) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *UserRepository) CountByRole(role models.UserRole) (int64, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	return r.coll.CountDocuments(ctx, bson.M{"role": role})
}

// FindAdmin returns the (single) admin user. Useful for sending admin alerts.
// Returns ErrNotFound when the system has no admin yet.
func (r *UserRepository) FindAdmin() (*models.User, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var user models.User
	err := r.coll.FindOne(ctx, bson.M{"role": models.RoleAdmin}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
