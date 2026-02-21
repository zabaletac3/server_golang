package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/eren_dev/go_server/internal/config"
)

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	timeout  time.Duration
}

func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.MongoTimeout)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return &MongoDB{
		client:   client,
		database: client.Database(cfg.MongoDatabase),
		timeout:  cfg.MongoTimeout,
	}, nil
}

func (m *MongoDB) DB() *mongo.Database {
	return m.database
}

func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

func (m *MongoDB) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()
	return m.client.Ping(ctx, readpref.Primary())
}

func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

func (m *MongoDB) Database() *mongo.Database {
	return m.database
}
