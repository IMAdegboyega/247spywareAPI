package repository

import (
	"context"
	"errors"
	"regexp"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostRepository struct {
	db   *config.MongoDB
	coll *mongo.Collection
}

func NewPostRepository(db *config.MongoDB) *PostRepository {
	return &PostRepository{db: db, coll: db.Database.Collection("posts")}
}

func (r *PostRepository) Create(post *models.Post) error {
	ctx, cancel := reqCtx()
	defer cancel()

	id, err := r.db.NextSequence(ctx, "posts")
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	post.ID = id
	post.CreatedAt = now
	post.UpdatedAt = now
	if post.Status == "" {
		post.Status = models.StatusDraft
	}

	_, err = r.coll.InsertOne(ctx, post)
	return err
}

func (r *PostRepository) FindByID(id uint) (*models.Post, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var post models.Post
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) FindBySlug(slug string) (*models.Post, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	var post models.Post
	err := r.coll.FindOne(ctx, bson.M{"slug": slug}).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) findPaginated(ctx context.Context, filter bson.M, sort bson.D, page, perPage int) ([]models.Post, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * perPage)
	opts := options.Find().
		SetSort(sort).
		SetSkip(skip).
		SetLimit(int64(perPage))

	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, total, err
	}
	defer cur.Close(ctx)

	var posts []models.Post
	if err := cur.All(ctx, &posts); err != nil {
		return nil, total, err
	}
	if posts == nil {
		posts = []models.Post{}
	}
	return posts, total, nil
}

func (r *PostRepository) FindAll(page, perPage int) ([]models.Post, int64, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	return r.findPaginated(ctx, bson.M{}, bson.D{{Key: "created_at", Value: -1}}, page, perPage)
}

func (r *PostRepository) FindPublished(page, perPage int) ([]models.Post, int64, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	return r.findPaginated(
		ctx,
		bson.M{"status": models.StatusPublished},
		bson.D{{Key: "published_at", Value: -1}},
		page, perPage,
	)
}

func (r *PostRepository) FindLatestNews(limit int) ([]models.Post, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	filter := bson.M{"status": models.StatusPublished, "is_latest_news": true}
	opts := options.Find().
		SetSort(bson.D{{Key: "latest_news_at", Value: -1}}).
		SetLimit(int64(limit))

	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var posts []models.Post
	if err := cur.All(ctx, &posts); err != nil {
		return nil, err
	}
	if posts == nil {
		posts = []models.Post{}
	}
	return posts, nil
}

func (r *PostRepository) FindByCategory(categoryID uint, page, perPage int) ([]models.Post, int64, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	return r.findPaginated(
		ctx,
		bson.M{"category_id": categoryID, "status": models.StatusPublished},
		bson.D{{Key: "published_at", Value: -1}},
		page, perPage,
	)
}

func (r *PostRepository) FindByAuthor(authorID uint, page, perPage int) ([]models.Post, int64, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	return r.findPaginated(
		ctx,
		bson.M{"author_id": authorID},
		bson.D{{Key: "created_at", Value: -1}},
		page, perPage,
	)
}

func (r *PostRepository) FindScheduledDue() ([]models.Post, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	now := time.Now().UTC()
	cur, err := r.coll.Find(ctx, bson.M{
		"status":         models.StatusScheduled,
		"scheduled_for":  bson.M{"$lte": now},
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var posts []models.Post
	if err := cur.All(ctx, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) FindExpiredLatestNews(duration time.Duration) ([]models.Post, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	cutoff := time.Now().UTC().Add(-duration)
	cur, err := r.coll.Find(ctx, bson.M{
		"is_latest_news":  true,
		"latest_news_at":  bson.M{"$lt": cutoff},
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var posts []models.Post
	if err := cur.All(ctx, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) Update(post *models.Post) error {
	ctx, cancel := reqCtx()
	defer cancel()

	post.UpdatedAt = time.Now().UTC()
	_, err := r.coll.ReplaceOne(ctx, bson.M{"_id": post.ID}, post)
	return err
}

func (r *PostRepository) Delete(id uint) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *PostRepository) IncrementViewCount(id uint) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"view_count": 1}},
	)
	return err
}

func (r *PostRepository) CountByStatus(status models.PostStatus) (int64, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	return r.coll.CountDocuments(ctx, bson.M{"status": status})
}

func (r *PostRepository) CountAll() (int64, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	return r.coll.CountDocuments(ctx, bson.M{})
}

// SumViewCount uses a small aggregation to total view_count across all posts.
func (r *PostRepository) SumViewCount() (int64, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{"_id": nil, "total": bson.M{"$sum": "$view_count"}}}},
	}
	cur, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cur.Close(ctx)

	if !cur.Next(ctx) {
		return 0, nil
	}
	var out struct {
		Total int64 `bson:"total"`
	}
	if err := cur.Decode(&out); err != nil {
		return 0, err
	}
	return out.Total, nil
}

func (r *PostRepository) Search(query string, page, perPage int) ([]models.Post, int64, error) {
	ctx, cancel := reqCtx()
	defer cancel()

	// Case-insensitive substring match on title or content. We escape the query
	// so user input can't inject regex meta-characters.
	pattern := regexp.QuoteMeta(query)
	rx := bson.M{"$regex": pattern, "$options": "i"}
	filter := bson.M{
		"status": models.StatusPublished,
		"$or": []bson.M{
			{"title": rx},
			{"content": rx},
		},
	}
	return r.findPaginated(
		ctx, filter,
		bson.D{{Key: "published_at", Value: -1}},
		page, perPage,
	)
}

// PropagateAuthorRename updates the embedded author summary across every post
// authored by this user. Called when a user's display name changes.
func (r *PostRepository) PropagateAuthorRename(authorID uint, newDisplayName string) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.UpdateMany(
		ctx,
		bson.M{"author_id": authorID},
		bson.M{"$set": bson.M{
			"author.display_name": newDisplayName,
			"updated_at":          time.Now().UTC(),
		}},
	)
	return err
}

// PropagateCategoryRename updates the embedded category summary across every
// post in the given category.
func (r *PostRepository) PropagateCategoryRename(categoryID uint, newName, newSlug string) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.UpdateMany(
		ctx,
		bson.M{"category_id": categoryID},
		bson.M{"$set": bson.M{
			"category.name": newName,
			"category.slug": newSlug,
			"updated_at":    time.Now().UTC(),
		}},
	)
	return err
}

// ClearCategoryFromPosts removes the embedded category and the id reference
// when a category is deleted. Matches the "posts become uncategorised" UX.
func (r *PostRepository) ClearCategoryFromPosts(categoryID uint) error {
	ctx, cancel := reqCtx()
	defer cancel()

	_, err := r.coll.UpdateMany(
		ctx,
		bson.M{"category_id": categoryID},
		bson.M{
			"$unset": bson.M{"category_id": "", "category": ""},
			"$set":   bson.M{"updated_at": time.Now().UTC()},
		},
	)
	return err
}
