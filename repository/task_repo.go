package repository

import (
	"context"
	"log"

	"github.com/tieubaoca/chatbot-be/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TaskRepo interface {
	CreateTask(ctx context.Context, task *types.Task) error
	GetTask(ctx context.Context, id string) (*types.Task, error)
	ListTasks(ctx context.Context, assignee, reporter string, status []string, createFromTime, deadline int64, limit, offset int) ([]*types.Task, error)
	UpdateTask(ctx context.Context, task *types.Task) error
	DeleteTask(ctx context.Context, id string) error
}

type taskRepo struct {
	collection *mongo.Collection
}

func NewTaskRepo(db *mongo.Database) TaskRepo {
	// check if collection does not exist, create one
	collectionNames, err := db.ListCollectionNames(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	collectionExists := false
	for _, name := range collectionNames {
		if name == "tasks" {
			collectionExists = true
			break
		}
	}
	collection := db.Collection("tasks")
	if !collectionExists {
		indexes := []mongo.IndexModel{

			{
				Keys: bson.D{
					{Key: "status", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "created_at", Value: -1},
					{Key: "deadline", Value: -1},
				},
			}}

		_, err = collection.Indexes().CreateMany(context.Background(), indexes)
		if err != nil {
			log.Printf("Error creating indexes: %v", err)
			return nil
		}
	}

	return &taskRepo{
		collection: db.Collection("tasks"),
	}
}

func (r *taskRepo) CreateTask(ctx context.Context, task *types.Task) error {
	_, err := r.collection.InsertOne(ctx, task)
	return err
}

func (r *taskRepo) GetTask(ctx context.Context, id string) (*types.Task, error) {
	var task types.Task
	err := r.collection.FindOne(ctx, map[string]string{"id": id}).Decode(&task)
	return &task, err
}

func (r *taskRepo) ListTasks(ctx context.Context, assignee, reporter string, status []string, createFromTime, deadline int64, limit, offset int) ([]*types.Task, error) {
	filter := make(map[string]interface{})
	if assignee != "" {
		filter["assignee"] = assignee
	}
	if reporter != "" {
		filter["reporter"] = reporter
	}
	if len(status) > 0 {
		filter["status"] = map[string]interface{}{"$in": status}
	}
	if createFromTime > 0 {
		filter["created_at"] = map[string]interface{}{"$gte": createFromTime}
	}
	if deadline > 0 {
		filter["deadline"] = map[string]interface{}{"$lte": deadline}
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []*types.Task
	for cursor.Next(ctx) {
		var task types.Task
		if err := cursor.Decode(&task); err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}
	return tasks, nil
}

func (r *taskRepo) UpdateTask(ctx context.Context, task *types.Task) error {
	_, err := r.collection.ReplaceOne(ctx, map[string]string{"id": task.ID}, task)
	return err
}

func (r *taskRepo) DeleteTask(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, map[string]string{"id": id})
	return err
}
