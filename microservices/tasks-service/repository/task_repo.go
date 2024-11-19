package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os"
	"tasks-service/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TaskRepo struct {
	Cli *mongo.Client
}

func NewTaskRepo(ctx context.Context, logger *log.Logger) (*TaskRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")
	if dburi == "" {
		return nil, fmt.Errorf("MONGO_DB_URI is not set")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Printf("MongoDB ping failed: %v", err)
		return nil, err
	}

	log.Println("Connected to MongoDB successfully")

	if err := insertInitialTasks(client); err != nil {
		log.Printf("Failed to insert initial tasks: %v", err)
	}

	return &TaskRepo{Cli: client}, nil
}

func (tr *TaskRepo) Disconnect(ctx context.Context) error {
	err := tr.Cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func insertInitialTasks(client *mongo.Client) error {
	collection := client.Database("mongoDemo").Collection("tasks")
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		log.Println("Error checking task count:", err)
		return err
	}

	if count > 0 {
		log.Println("Tasks already exist in the database")
		return nil
	}

	// Define initial tasks to insert
	tasks := []interface{}{
		domain.Task{
			Name:        "Task 1",
			Description: "This is the first task.",
			Status:      domain.Status(0),
			ProjectID:   "jnasdndslksad",
		},
		domain.Task{
			Name:        "Task 2",
			Description: "This is the second task.",
			Status:      domain.Status(0),
			ProjectID:   "lksaddsmamkls",
		},
	}

	// Insert initial tasks
	_, err = collection.InsertMany(context.Background(), tasks)
	if err != nil {
		log.Println("Error inserting initial tasks:", err)
		return err
	}

	log.Println("Initial tasks inserted successfully")
	return nil
}

func (tr *TaskRepo) getCollection() *mongo.Collection {
	if tr.Cli == nil {
		log.Println("Mongo client is nil!")
		return nil
	}

	if err := tr.Cli.Ping(context.Background(), nil); err != nil {
		log.Println("Error pinging MongoDB, connection lost:", err)
		return nil
	}

	return tr.Cli.Database("mongoDemo").Collection("tasks")
}

func (tr *TaskRepo) Create(task domain.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println(task)

	collection := tr.getCollection()
	if collection == nil {
		log.Println("Failed to retrieve collection")
		return fmt.Errorf("collection is nil")
	}

	_, err := collection.InsertOne(ctx, task)
	if err != nil {
		log.Println("Error inserting task:", err, task)
		return err
	}

	log.Println("Task created successfully:", task)
	return nil
}

func (tr *TaskRepo) GetAll() ([]domain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return nil, fmt.Errorf("failed to retrieve collection")
	}

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Println("Error finding tasks:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []domain.Task
	for cursor.Next(ctx) {
		var task domain.Task
		if err := cursor.Decode(&task); err != nil {
			log.Println("Error decoding task:", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := cursor.Err(); err != nil {
		log.Println("Error iterating over cursor:", err)
		return nil, err
	}

	return tasks, nil
}

func (tr *TaskRepo) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	taskCollection := tr.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	result, err := taskCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Documents deleted: %v\n", result.DeletedCount)
	return nil
}

func (tr *TaskRepo) DeleteAllByProjectID(projectID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return fmt.Errorf("failed to retrieve collection")
	}

	filter := bson.M{"project_id": projectID}
	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Println("Error deleting tasks by ProjectID:", err)
		return err
	}

	log.Printf("Tasks with ProjectID %s deleted successfully", projectID)
	return nil
}

func (tr *TaskRepo) GetAllByProjectID(projectID string) (domain.Tasks, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tasksCollection := tr.getCollection()
	var tasks domain.Tasks

	// Query only projects where the manager's ID matches the provided id
	filter := bson.M{"project_id": projectID}
	cursor, err := tasksCollection.Find(ctx, filter)
	if err != nil {
		log.Println("Error finding tasks:", err)
		return nil, err

	}

	if err := cursor.Err(); err != nil {
		log.Println("Error iterating over tasks:", err)
		return nil, err
	}

	log.Printf("Fetched %d tasks with ProjectID %s", len(tasks), projectID)
	return tasks, nil
}

func (tr *TaskRepo) GetById(id string) (*domain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := tr.getCollection()

	// Convert id string to ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Invalid ID format:", err)
		return nil, err
	}

	// Find project by _id
	filter := bson.M{"_id": objID}
	var t domain.Task
	err = projectsCollection.FindOne(ctx, filter).Decode(&t)
	if err != nil {
		log.Println("Error finding task by ID:", err)
		return nil, err
	}

	return &t, nil
}
