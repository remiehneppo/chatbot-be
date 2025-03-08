package repository

import (
	"context"

	"github.com/tieubaoca/chatbot-be/types"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *types.User) error
	GetUser(ctx context.Context, id string) (*types.User, error)
	GetUserByWorkspace(ctx context.Context, workspace string) ([]*types.User, error)
	UpdateUser(ctx context.Context, user *types.User) error
	DeleteUser(ctx context.Context, id string) error
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

func (r *userRepo) GetUser(ctx context.Context, id string) (*types.User, error) {
	var user types.User
	err := r.collection.FindOne(ctx, map[string]string{"id": id}).Decode(&user)
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

func (r *userRepo) UpdateUser(ctx context.Context, user *types.User) error {
	_, err := r.collection.ReplaceOne(ctx, map[string]string{"id": user.ID}, user)
	return err
}

func (r *userRepo) DeleteUser(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, map[string]string{"id": id})
	return err
}
