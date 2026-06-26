package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB wraps the client + active database handle. Pass it into repositories.
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// InitDB connects to MongoDB using MONGODB_URI and selects the database named
// in MONGODB_DB. It also runs index setup and seeds the counters collection.
func InitDB() (*MongoDB, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, errors.New("MONGODB_URI is not set")
	}

	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = "247techspyware"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri).SetServerSelectionTimeout(20 * time.Second)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	db := &MongoDB{
		Client:   client,
		Database: client.Database(dbName),
	}

	if err := db.ensureIndexes(ctx); err != nil {
		return nil, fmt.Errorf("ensure indexes: %w", err)
	}

	log.Printf("Connected to MongoDB database %q", dbName)
	return db, nil
}

// ensureIndexes creates the indexes the app relies on. Safe to run repeatedly.
func (m *MongoDB) ensureIndexes(ctx context.Context) error {
	indexes := map[string][]mongo.IndexModel{
		"users": {
			{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "role", Value: 1}}},
		},
		"categories": {
			{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"posts": {
			{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "status", Value: 1}}},
			{Keys: bson.D{{Key: "author_id", Value: 1}}},
			{Keys: bson.D{{Key: "category_id", Value: 1}}},
			{Keys: bson.D{{Key: "is_latest_news", Value: 1}, {Key: "latest_news_at", Value: -1}}},
			{Keys: bson.D{{Key: "published_at", Value: -1}}},
			{Keys: bson.D{{Key: "scheduled_for", Value: 1}}},
		},
		"editor_picks": {
			{Keys: bson.D{{Key: "post_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "order", Value: 1}}},
		},
		"subscribers": {
			{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "unsubscribe_token", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"adverts": {
			{Keys: bson.D{{Key: "is_active", Value: 1}}},
			{Keys: bson.D{{Key: "position", Value: 1}}},
		},
	}

	for coll, models := range indexes {
		if _, err := m.Database.Collection(coll).Indexes().CreateMany(ctx, models); err != nil {
			return fmt.Errorf("create indexes on %s: %w", coll, err)
		}
	}

	return nil
}

// NextSequence atomically increments and returns the next id for the given
// counter name. The counters collection holds documents of the form
// `{_id: "<name>", seq: <int>}`. We start at 1.
func (m *MongoDB) NextSequence(ctx context.Context, name string) (uint, error) {
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result struct {
		Seq uint `bson:"seq"`
	}
	err := m.Database.Collection("counters").FindOneAndUpdate(
		ctx,
		bson.M{"_id": name},
		bson.M{"$inc": bson.M{"seq": uint(1)}},
		opts,
	).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("next sequence for %s: %w", name, err)
	}
	return result.Seq, nil
}

// Disconnect gracefully closes the client. Call from main on shutdown.
func (m *MongoDB) Disconnect(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}
