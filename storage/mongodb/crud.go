package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// inserts a single document.
func (m *Mongo) InsertOne(
	ctx context.Context,
	db, coll string,
	doc interface{},
	opts ...*options.InsertOneOptions,
) (*mongo.InsertOneResult, error) {
	return m.Collection(db, coll).InsertOne(ctx, doc, opts...)
}

// retrieves the first matching document into result.
func (m *Mongo) FindOne(
	ctx context.Context,
	db, coll string,
	filter interface{},
	result interface{},
	opts ...*options.FindOneOptions,
) error {
	return m.Collection(db, coll).FindOne(ctx, filter, opts...).Decode(result)
}

// updates the first document matching the filter.
func (m *Mongo) UpdateOne(
	ctx context.Context,
	db, coll string,
	filter, update interface{},
	opts ...*options.UpdateOptions,
) (*mongo.UpdateResult, error) {
	return m.Collection(db, coll).UpdateOne(ctx, filter, update, opts...)
}
