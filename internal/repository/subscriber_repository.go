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

type SubscriberRepository struct {
	db   *config.MongoDB
	coll *mongo.Collection
}

func NewSubscriberRepository(db *config.MongoDB) *SubscriberRepository {
	return &SubscriberRepository{db: db, coll: db.Database.Collection("subscribers")}
}

func (r *SubscriberRepository) Create(sub *models.Subscriber) error {
	ctx, cancel := reqCtx()
	defer cancel()

	id, err := r.db.NextSequence(ctx, "subscribers")
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	sub.ID = id
	sub.CreatedAt = now
	sub.UpdatedAt = now
	if sub.SubscribedAt.IsZero() {
		sub.SubscribedAt = now
	}

	_, err = r.coll.InsertOne(ctx, sub)
	return err
}

func (r *SubscriberRepository) FindByEmail(email string) (*models.Subscriber, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var sub models.Subscriber
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&sub)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriberRepository) FindByToken(token string) (*models.Subscriber, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var sub models.Subscriber
	err := r.coll.FindOne(ctx, bson.M{"unsubscribe_token": token}).Decode(&sub)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriberRepository) FindByID(id uint) (*models.Subscriber, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var sub models.Subscriber
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&sub)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriberRepository) FindAllActive() ([]models.Subscriber, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "subscribed_at", Value: -1}})
	cur, err := r.coll.Find(ctx, bson.M{"is_active": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var subs []models.Subscriber
	if err := cur.All(ctx, &subs); err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *SubscriberRepository) FindAll() ([]models.Subscriber, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "subscribed_at", Value: -1}})
	cur, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var subs []models.Subscriber
	if err := cur.All(ctx, &subs); err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *SubscriberRepository) Update(sub *models.Subscriber) error {
	ctx, cancel := reqCtx()
	defer cancel()

	sub.UpdatedAt = time.Now().UTC()
	_, err := r.coll.ReplaceOne(ctx, bson.M{"_id": sub.ID}, sub)
	return err
}

func (r *SubscriberRepository) Delete(id uint) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *SubscriberRepository) CountActive() (int64, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	return r.coll.CountDocuments(ctx, bson.M{"is_active": true})
}
