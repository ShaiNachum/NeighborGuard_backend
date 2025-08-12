package services

import (
	"context"
	"errors"
	"neighborguard/pkg/database"
	"sort"
	"time"
	"fmt"

	"github.com/umahmood/haversine"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)



type Gender string
type Role string
type MeetingAssistanceStatus string

const (
	Male   Gender = "MALE"
	Female Gender = "FEMALE"
)

const (
	Volunteer Role = "VOLUNTEER"
	Recipient Role = "RECIPIENT"
)

const (
	DoNotNeedAssistance MeetingAssistanceStatus = "DO_NOT_NEED_ASSISTANCE"
	NeedAssistance      MeetingAssistanceStatus = "NEED_ASSISTANCE"
	InProgress          MeetingAssistanceStatus = "IN_PROGRESS"
	Provide             MeetingAssistanceStatus = "PROVIDE"
	DoNotProvide        MeetingAssistanceStatus = "DO_NOT_PROVIDE"
)

type LonLat struct {
	Longitude float64 `json:"longitude" bson:"longitude"`
	Latitude  float64 `json:"latitude" bson:"latitude"`
}

type Address struct {
	City            string `json:"city" bson:"city"`
	Street          string `json:"street" bson:"street"`
	HouseNumber     int    `json:"houseNumber" bson:"houseNumber"`
	ApartmentNumber int    `json:"apartmentNumber" bson:"apartmentNumber"`
}

type NewUser struct {
	FirstName    string                             `json:"firstName"`
	LastName     string                             `json:"lastName"`
	Age          int                                `json:"age"`
	PhoneNumber  string                             `json:"phoneNumber"`
	Gender       Gender                             `json:"gender"`
	Email        string                             `json:"email"`
	Password     string                             `json:"password"`
	Address      Address                            `json:"address"`
	Languages    []string                           `json:"languages"`
	Services     map[string]MeetingAssistanceStatus `json:"services"` //map[ServiceName]MeetingAssistanceStatus
	Role         Role                               `json:"role"`
	LonLat       LonLat                             `json:"lonLat"`
	LastOK       int64                              `json:"lastOK"`
	ProfileImage string                             `json:"profileImage"`
}

type User struct {
	ID           string                             `json:"uid" bson:"_id,omitempty"`
	FirstName    string                             `json:"firstName" bson:"firstName"`
	LastName     string                             `json:"lastName" bson:"lastName"`
	Age          int                                `json:"age" bson:"age"`
	PhoneNumber  string                             `json:"phoneNumber" bson:"phoneNumber"`
	Gender       Gender                             `json:"gender" bson:"gender"`
	Email        string                             `json:"email" bson:"email"`
	Password     string                             `json:"password" bson:"password"`
	Address      Address                            `json:"address" bson:"address"`
	Languages    []string                           `json:"languages" bson:"languages"`
	Services     map[string]MeetingAssistanceStatus `json:"services" bson:"services"`
	Role         Role                               `json:"role" bson:"role"`
	LonLat       LonLat                             `json:"lonLat" bson:"lonLat"`
	LastOK       int64                              `json:"lastOK" bson:"lastOK"`
	ProfileImage string                             `json:"profileImage" bson:"profileImage"`
	CreatedAt    time.Time                          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time                          `json:"updatedAt" bson:"updatedAt"`
}


func GetNearbyRecipients(
	volunteerUID string,
	filterByLat *float64,
	filterByLon *float64,
) ([]User, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the volunteer's details from MongoDB
	var volunteer User
	err := database.UsersCollection.FindOne(ctx, bson.M{"_id": volunteerUID}).Decode(&volunteer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("volunteer not found")
		}
		return nil, err
	}

	// Verify that the user is actually a volunteer
	if volunteer.Role != Volunteer {
		return nil, errors.New("only volunteers can use this endpoint")
	}

	// Find all recipients in the database
	cursor, err := database.UsersCollection.Find(ctx, bson.M{"role": string(Recipient)})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Load all recipients into memory
	var recipients []User
	if err = cursor.All(ctx, &recipients); err != nil {
		return nil, err
	}

	// Apply additional filtering criteria in memory
	var filtered []User
	for _, user := range recipients {
		// Filter by location if coordinates are provided
		if filterByLat != nil && filterByLon != nil {
			nearLocation := LonLat{Longitude: *filterByLon, Latitude: *filterByLat}
			if !isInLocation(user.LonLat, nearLocation) {
				continue
			}
		}

		// Check for language match
		if !hasCommonElements(volunteer.Languages, user.Languages) {
			continue
		}

		// Check for service matching and assistance need
		if !checkAssistanceAndServices(user, volunteer) {
			continue
		}

		filtered = append(filtered, user)
	}

	// Sort recipients by priority (General Check needs and LastOK time)
	sort.Slice(filtered, func(i, j int) bool {
		recipientI := filtered[i]
		recipientJ := filtered[j]

		// Check if either recipient needs General Check
		needsGeneralCheckI := recipientI.Services["General Check"] == NeedAssistance
		needsGeneralCheckJ := recipientJ.Services["General Check"] == NeedAssistance

		// If both have the same General Check status, sort by LastOK time
		if needsGeneralCheckI == needsGeneralCheckJ {
			return recipientI.LastOK < recipientJ.LastOK
		}

		// Prioritize recipients who need General Check
		return needsGeneralCheckI
	})

	return filtered, nil
}


func CreateUser(newUser NewUser) (User, error) {
	// Create a context with timeout for database operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if email already exists to prevent duplicates
	existingFilter := bson.M{"email": newUser.Email}
	count, err := database.UsersCollection.CountDocuments(ctx, existingFilter)
	if err != nil {
		return User{}, err
	}
	if count > 0 {
		return User{}, errors.New("user with this email already exists")
	}

	// Get current time for timestamps
	now := time.Now()
	
	// Create a new user with a MongoDB ObjectID
	user := User{
		ID:           primitive.NewObjectID().Hex(), // Generate a new MongoDB ObjectID
		FirstName:    newUser.FirstName,
		LastName:     newUser.LastName,
		Age:          newUser.Age,
		PhoneNumber:  newUser.PhoneNumber,
		Gender:       newUser.Gender,
		Email:        newUser.Email,
		Password:     newUser.Password,  // Note: In production, passwords should be hashed
		Address:      newUser.Address,
		Languages:    newUser.Languages,
		Services:     newUser.Services,
		Role:         newUser.Role,
		LonLat:       newUser.LonLat,
		LastOK:       now.Unix(),
		ProfileImage: newUser.ProfileImage,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Insert the user document into MongoDB
	_, err = database.UsersCollection.InsertOne(ctx, user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}


func UpdateUser(uid string, updatedUser User) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify the user exists before updating
	var existingUser User
	err := database.UsersCollection.FindOne(ctx, bson.M{"_id": uid}).Decode(&existingUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("user not found")
		}
		return err
	}

	// Create an update operation with the fields to be modified
	update := bson.M{
		"$set": bson.M{
			"firstName":    updatedUser.FirstName,
			"lastName":     updatedUser.LastName,
			"phoneNumber":  updatedUser.PhoneNumber,
			"languages":    updatedUser.Languages,
			"services":     updatedUser.Services,
			"address":      updatedUser.Address,
			"lonLat":       updatedUser.LonLat,
			"lastOK":       updatedUser.LastOK,
			"profileImage": updatedUser.ProfileImage,
			"updatedAt":    time.Now(),
		},
	}

	// Execute the update in MongoDB
	_, err = database.UsersCollection.UpdateOne(ctx, bson.M{"_id": uid}, update)
	return err
}

func GetUserByEmail(email string) (*User, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query MongoDB for a user with the specified email
	var user User
	err := database.UsersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// Helper functions:

// Helper function to check if two string slices have any common elements
func hasCommonElements(slice1, slice2 []string) bool {
	set := make(map[string]bool)
	for _, item := range slice1 {
		set[item] = true
	}
	for _, item := range slice2 {
		if set[item] {
			return true
		}
	}
	return false
}

// Helper function to check if a user is within 1km of a location
func isInLocation(userLonLat LonLat, nearLocation LonLat) bool {
	point1 := haversine.Coord{Lat: userLonLat.Latitude, Lon: userLonLat.Longitude}
	point2 := haversine.Coord{Lat: nearLocation.Latitude, Lon: nearLocation.Longitude}

	_, km := haversine.Distance(point1, point2)
	return km <= 1.0
}

func checkAssistanceAndServices(recipient User, volunteer User) bool {
    // Initialize services map if nil
    if recipient.Services == nil {
        recipient.Services = make(map[string]MeetingAssistanceStatus)
    }

    // Print recipient's services for debugging
    fmt.Printf("Checking services for recipient %s: %+v\n", recipient.ID, recipient.Services)

    // Check for time-based general check need
    timeBasedNeed := time.Now().Unix() - recipient.LastOK > 60 // 1 minute for testing
    
    // Track if services were updated
    updated := false
    
    // If time-based need detected, update General Check status in MongoDB
    if timeBasedNeed && recipient.Services["General Check"] != NeedAssistance && recipient.Services["General Check"] != InProgress {
        fmt.Printf("Recipient %s needs General Check (time-based)\n", recipient.ID)
        
        // Update the recipient's General Check service in MongoDB
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        // Set the General Check service to NeedAssistance
        update := bson.M{"$set": bson.M{"services.General Check": string(NeedAssistance)}}
        
        // Update the document in MongoDB
        result, err := database.UsersCollection.UpdateOne(ctx, bson.M{"_id": recipient.ID}, update)
        if err != nil {
            // Log the error but continue processing
            fmt.Printf("Error updating General Check status: %v\n", err)
        } else {
            fmt.Printf("Updated General Check status for recipient %s: %+v\n", recipient.ID, result)
            updated = true
        }
        
        // Also update our local copy of the recipient for this function
        recipient.Services["General Check"] = NeedAssistance
    }

    // Check for any service in NEED_ASSISTANCE that matches volunteer's provided services
    hasMatchingService := false
    for service, status := range recipient.Services {
        // Skip services already in progress
        if status == InProgress {
            fmt.Printf("Service %s is already in progress for recipient %s\n", service, recipient.ID)
            continue
        }

        // Check if this service needs assistance and volunteer can provide
        if status == NeedAssistance && volunteer.Services[service] == Provide {
            fmt.Printf("Found matching service %s for recipient %s\n", service, recipient.ID)
            hasMatchingService = true
            break
        }
    }

    // If we updated the recipient's services but didn't find a matching service yet, 
    // reload the recipient from the database to get the freshest data
    if updated && !hasMatchingService {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        var updatedRecipient User
        err := database.UsersCollection.FindOne(ctx, bson.M{"_id": recipient.ID}).Decode(&updatedRecipient)
        if err == nil {
            // Check again with the fresh data
            for service, status := range updatedRecipient.Services {
                if status == NeedAssistance && volunteer.Services[service] == Provide {
                    fmt.Printf("Found matching service %s for recipient %s after refresh\n", service, recipient.ID)
                    hasMatchingService = true
                    break
                }
            }
        }
    }

    return hasMatchingService
}
