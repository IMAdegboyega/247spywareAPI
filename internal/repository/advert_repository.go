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

type AdvertRepository struct {
	db   *config.MongoDB
	coll *mongo.Collection
}

func NewAdvertRepository(db *config.MongoDB) *AdvertRepository {
	return &AdvertRepository{db: db, coll: db.Database.Collection("adverts")}
}

func (r *AdvertRepository) Create(advert *models.Advert) error {
	ctx, cancel := reqCtx()
	defer cancel()

	id, err := r.db.NextSequence(ctx, "adverts")
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	advert.ID = id
	advert.CreatedAt = now
	advert.UpdatedAt = now

	_, err = r.coll.InsertOne(ctx, advert)
	return err
}

func (r *AdvertRepository) FindByID(id uint) (*models.Advert, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var ad models.Advert
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&ad)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &ad, nil
}

func (r *AdvertRepository) FindAll() ([]models.Advert, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var ads []models.Advert
	if err := cur.All(ctx, &ads); err != nil {
		return nil, err
	}
	return ads, nil
}

// activeFilter returns the filter for ads that are currently runnable.
func activeFilter(now time.Time) bson.M {
	return bson.M{
		"is_active": true,
		"$and": []bson.M{
			{"$or": []bson.M{
				{"start_date": bson.M{"$exists": false}},
				{"start_date": nil},
				{"start_date": bson.M{"$lte": now}},
			}},
			{"$or": []bson.M{
				{"end_date": bson.M{"$exists": false}},
				{"end_date": nil},
				{"end_date": bson.M{"$gte": now}},
			}},
		},
	}
}

func (r *AdvertRepository) FindActive() ([]models.Advert, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	cur, err := r.coll.Find(ctx, activeFilter(time.Now().UTC()))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var ads []models.Advert
	if err := cur.All(ctx, &ads); err != nil {
		return nil, err
	}
	return ads, nil
}

func (r *AdvertRepository) FindByPosition(position string) ([]models.Advert, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	filter := activeFilter(time.Now().UTC())
	filter["position"] = position

	cur, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var ads []models.Advert
	if err := cur.All(ctx, &ads); err != nil {
		return nil, err
	}
	return ads, nil
}

func (r *AdvertRepository) Update(advert *models.Advert) error {
	ctx, cancel := reqCtx()
	defer cancel()

	advert.UpdatedAt = time.Now().UTC()
	_, err := r.coll.ReplaceOne(ctx, bson.M{"_id": advert.ID}, advert)
	return err
}

func (r *AdvertRepository) Delete(id uint) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *AdvertRepository) IncrementClickCount(id uint) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"click_count": 1}})
	return err
}

func (r *AdvertRepository) IncrementViewCount(id uint) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"view_count": 1}})
	return err
}
