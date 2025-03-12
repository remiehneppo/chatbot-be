package repository

import (
	"context"

	"github.com/tieubaoca/chatbot-be/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AdminRepo interface {
	GetAdmin(ctx context.Context, id string) (*types.Admin, error)
	GetAdminByUsername(ctx context.Context, username string) (*types.Admin, error)
	CreateAdmin(ctx context.Context, admin *types.Admin) error
	UpdateAdmin(ctx context.Context, id string, admin *types.Admin) error
	DeleteAdmin(ctx context.Context, id string) error
}

type adminRepo struct {
	collection *mongo.Collection
}

func NewAdminRepo(collection *mongo.Collection) AdminRepo {
	return &adminRepo{
		collection: collection,
	}
}

func (r *adminRepo) GetAdmin(ctx context.Context, id string) (*types.Admin, error) {
	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	admin := &types.Admin{}
	if err := r.collection.FindOne(ctx, bson.M{"_id": objId}).Decode(admin); err != nil {
		return nil, err
	}
	return admin, nil
}

func (r *adminRepo) CreateAdmin(ctx context.Context, admin *types.Admin) error {
	_, err := r.collection.InsertOne(ctx, admin)
	return err
}

func (r *adminRepo) UpdateAdmin(ctx context.Context, id string, admin *types.Admin) error {
	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": admin})
	return err
}

func (r *adminRepo) DeleteAdmin(ctx context.Context, id string) error {
	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objId})
	return err
}

func (r *adminRepo) GetAdminByUsername(ctx context.Context, username string) (*types.Admin, error) {
	admin := &types.Admin{}
	if err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(admin); err != nil {
		return nil, err
	}
	return admin, nil
}
