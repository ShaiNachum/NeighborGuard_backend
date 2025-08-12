package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB client and collections
var (
	Client          *mongo.Client
	UsersCollection *mongo.Collection
	MeetingsCollection *mongo.Collection
)

// Connect establishes a connection to MongoDB
func Connect() error {
	// Get MongoDB URI from environment variable or use default for local development
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		// Default connection string for local development
		// In production, use environment variables instead
		mongoURI = "mongodb+srv://shainachum111:UBLi1TO9A37NY2ya@cluster0.fahnj.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
		log.Println("Using default MongoDB connection string. Consider setting MONGO_URI environment variable.")
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the MongoDB server to verify connection
	err = Client.Ping(ctx, readpref.Primary())
	if err != nil {
		return err
	}

	log.Println("Connected to MongoDB successfully!")

	// Get database and collections
	database := Client.Database("neighborguard")
	UsersCollection = database.Collection("users")
	MeetingsCollection = database.Collection("meetings")

	return nil
}

// Disconnect closes the MongoDB connection
func Disconnect() {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := Client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
		log.Println("Disconnected from MongoDB")
	}
}