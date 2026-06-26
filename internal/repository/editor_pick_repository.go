package repository

import (
	"errors"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type EditorPickRepository struct {
	db   *config.MongoDB
	coll *mongo.Collection
}

func NewEditorPickRepository(db *config.MongoDB) *EditorPickRepository {
	return &EditorPickRepository{db: db, coll: db.Database.Collection("editor_picks")}
}

func (r *EditorPickRepository) Create(pick *models.EditorPick) error {
	ctx, cancel := reqCtx()
	defer cancel()

	id, err := r.db.NextSequence(ctx, "editor_picks")
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	pick.ID = id
	pick.CreatedAt = now
	pick.UpdatedAt = now
	if pick.PickedAt.IsZero() {
		pick.PickedAt = now
	}
	// Don't persist the embedded Post pointer — that's only populated on read
	// via $lookup. Stash it locally and restore after the insert so the caller
	// keeps any in-memory data it had.
	embedded := pick.Post
	pick.Post = nil
	_, err = r.coll.InsertOne(ctx, pick)
	pick.Post = embedded
	return err
}

// findOneWithPost is a private helper that fetches a single pick + its post
// via $lookup so callers can render it without an extra round-trip.
func (r *EditorPickRepository) findOneWithPost(filter bson.M) (*models.EditorPick, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "posts",
			"localField":   "post_id",
			"foreignField": "_id",
			"as":           "post",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$post", "preserveNullAndEmptyArrays": true}}},
		{{Key: "$limit", Value: 1}},
	}

	cur, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if !cur.Next(ctx) {
		return nil, ErrNotFound
	}
	var pick models.EditorPick
	if err := cur.Decode(&pick); err != nil {
		return nil, err
	}
	return &pick, nil
}

func (r *EditorPickRepository) FindByID(id uint) (*models.EditorPick, error) {
	return r.findOneWithPost(bson.M{"_id": id})
}

func (r *EditorPickRepository) FindByPostID(postID uint) (*models.EditorPick, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var pick models.EditorPick
	err := r.coll.FindOne(ctx, bson.M{"post_id": postID}).Decode(&pick)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &pick, nil
}

// FindAll returns every pick joined with its post, ordered by `order` ascending
// then picked_at descending as a tiebreaker.
func (r *EditorPickRepository) FindAll() ([]models.EditorPick, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.M{
			"from":         "posts",
			"localField":   "post_id",
			"foreignField": "_id",
			"as":           "post",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$post", "preserveNullAndEmptyArrays": true}}},
		{{Key: "$sort", Value: bson.D{{Key: "order", Value: 1}, {Key: "picked_at", Value: -1}}}},
	}

	cur, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var picks []models.EditorPick
	if err := cur.All(ctx, &picks); err != nil {
		return nil, err
	}
	if picks == nil {
		picks = []models.EditorPick{}
	}
	return picks, nil
}

func (r *EditorPickRepository) Update(pick *models.EditorPick) error {
	ctx, cancel := reqCtx()
	defer cancel()

	pick.UpdatedAt = time.Now().UTC()
	embedded := pick.Post
	pick.Post = nil
	_, err := r.coll.ReplaceOne(ctx, bson.M{"_id": pick.ID}, pick)
	pick.Post = embedded
	return err
}

func (r *EditorPickRepository) Delete(id uint) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *EditorPickRepository) DeleteByPostID(postID uint) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.DeleteMany(ctx, bson.M{"post_id": postID})
	return err
}

// UpdateOrder writes back the order field for each pick in the slice.
// Wrapped in a transaction when the deployment is a replica set, falls back
// to a best-effort loop on standalone instances (Atlas free tier is a
// replica set so transactions are available).
func (r *EditorPickRepository) UpdateOrder(picks []models.EditorPick) error {
	ctx, cancel := reqCtx()
	defer cancel()

	session, err := r.db.Client.StartSession()
	if err != nil {
		// Fall back to a non-transactional loop.
		for _, p := range picks {
			if _, e := r.coll.UpdateOne(ctx, bson.M{"_id": p.ID}, bson.M{"$set": bson.M{"order": p.Order, "updated_at": time.Now().UTC()}}); e != nil {
				return e
			}
		}
		return nil
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		for _, p := range picks {
			if _, e := r.coll.UpdateOne(sc, bson.M{"_id": p.ID}, bson.M{"$set": bson.M{"order": p.Order, "updated_at": time.Now().UTC()}}); e != nil {
				return nil, e
			}
		}
		return nil, nil
	})
	return err
}

func (r *EditorPickRepository) Exists(postID uint) (bool, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	count, err := r.coll.CountDocuments(ctx, bson.M{"post_id": postID})
	return count > 0, err
}
