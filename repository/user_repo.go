package repository

import (
	"context"

	"github.com/tieubaoca/chatbot-be/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *types.User) error
	BatchCreateUser(ctx context.Context, users []*types.User) error
	GetUser(ctx context.Context, id string) (*types.User, error)
	GetUserByWorkspace(ctx context.Context, workspace string) ([]*types.User, error)
	PaginateUser(ctx context.Context, page int64, limit int64) ([]*types.User, int64, error)
	UpdateUser(ctx context.Context, id string, user *types.User) error
	DeleteUser(ctx context.Context, id string) error
	GetUserByUsername(ctx context.Context, username string) (*types.User, error)
}

type userRepo struct {
	collection *mongo.Collection
}

func NewUserRepo(collection *mongo.Collection) UserRepo {
	return &userRepo{
		collection: collection,
	}
}

func (r *userRepo) CreateUser(ctx context.Context, user *types.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *userRepo) BatchCreateUser(ctx context.Context, users []*types.User) error {
	var docs []interface{}
	for _, user := range users {
		docs = append(docs, user)
	}
	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

func (r *userRepo) GetUser(ctx context.Context, id string) (*types.User, error) {
	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var user types.User
	err = r.collection.FindOne(ctx, map[string]bson.ObjectID{"_id": objId}).Decode(&user)
	return &user, err
}

func (r *userRepo) GetUserByWorkspace(ctx context.Context, workspace string) ([]*types.User, error) {
	cursor, err := r.collection.Find(ctx, map[string]string{"workspace": workspace})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*types.User
	for cursor.Next(ctx) {
		var user types.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func (r *userRepo) PaginateUser(ctx context.Context, page int64, limit int64) ([]*types.User, int64, error) {
	opts := options.Find().SetSkip(page * limit).SetLimit(limit)
	cursor, err := r.collection.Find(ctx, nil, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []*types.User
	for cursor.Next(ctx) {
		var user types.User
		if err := cursor.Decode(&user); err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	total, err := r.collection.CountDocuments(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepo) UpdateUser(ctx context.Context, id string, user *types.User) error {
	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, map[string]bson.ObjectID{"_id": objId}, user)
	return err
}

func (r *userRepo) DeleteUser(ctx context.Context, id string) error {
	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, map[string]bson.ObjectID{"_id": objId})
	return err
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	var user types.User
	err := r.collection.FindOne(ctx, map[string]string{"username": username}).Decode(&user)
	return &user, err
}
