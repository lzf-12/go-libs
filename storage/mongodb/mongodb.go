package mongodb

import (
	"context"
	"errors"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Mongo wraps a *mongo.Client and exposes a few quality-of-life helpers.
type Mongo struct {
	client *mongo.Client
}

// Default connection-pool & timeout parameters.
// Tweak if your workload needs something markedly different.
const (
	defaultMaxPoolSize uint64 = 100
	defaultMinPoolSize uint64 = 0
	defaultConnTimeout        = 10 * time.Second
	defaultPingTimeout        = 5 * time.Second
)

func NewMongo(
	uri string,
	maxPool, minPool *uint64,
	connTimeout *time.Duration,
) (*Mongo, error) {

	if err := validateURI(uri); err != nil {
		return nil, err
	}

	mp := defaultMaxPoolSize
	if maxPool != nil {
		mp = *maxPool
	}

	mn := defaultMinPoolSize
	if minPool != nil {
		mn = *minPool
	}

	ct := defaultConnTimeout
	if connTimeout != nil {
		ct = *connTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), ct)
	defer cancel()

	opts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(mp).
		SetMinPoolSize(mn)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	return &Mongo{client: client}, nil
}

func (m *Mongo) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()
	return m.client.Ping(ctx, readpref.Primary())
}

// IsReady() alias for Ping()
func (m *Mongo) IsReady() error { return m.Ping() }

// Disconnect() closes every pooled connection.
func (m *Mongo) Disconnect(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// Client exposes the underlying *mongo.Client for advanced plumbing.
func (m *Mongo) Client() *mongo.Client { return m.client }

// Database returns a *mongo.Database handle.
func (m *Mongo) Database(name string, opts ...*options.DatabaseOptions) *mongo.Database {
	return m.client.Database(name, opts...)
}

// Collection is shorthand for m.Database(db).Collection(coll).
func (m *Mongo) Collection(
	db, coll string,
	collOpts ...*options.CollectionOptions,
) *mongo.Collection {
	return m.client.Database(db).Collection(coll, collOpts...)
}

// validate URI format
func validateURI(uri string) error {
	parsed, err := url.Parse(uri)
	if err != nil {
		return err
	}
	switch parsed.Scheme {
	case "mongodb", "mongodb+srv":
	default:
		return errors.New("URI must start with mongodb:// or mongodb+srv://")
	}
	if parsed.Host == "" {
		return errors.New("URI must include host")
	}
	return nil
}
