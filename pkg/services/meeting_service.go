package services

import (
	"context"
	"errors"
	"fmt"
	"neighborguard/pkg/database"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MeetingStatus string

const (
	IsPicked MeetingStatus = "IS_PICKED"
	Done     MeetingStatus = "DONE"
)

type NewMeeting struct {
	Recipient     User          `json:"recipient"`
	Volunteer     User          `json:"volunteer"`
	Date          int64         `json:"date"`
	Services      []string      `json:"services"` //list of services that will be provided on this meeting
	MeetingStatus MeetingStatus `json:"meetingStatus"`
}

type Meeting struct {
	ID            string        `json:"uid" bson:"_id,omitempty"`
	Recipient     User          `json:"recipient" bson:"-"`       // Not stored directly in MongoDB
	Volunteer     User          `json:"volunteer" bson:"-"`       // Not stored directly in MongoDB
	RecipientID   string        `json:"-" bson:"recipientId"`     // Store only the ID in MongoDB
	VolunteerID   string        `json:"-" bson:"volunteerId"`     // Store only the ID in MongoDB
	Date          int64         `json:"date" bson:"date"`
	Services      []string      `json:"services" bson:"services"`
	MeetingStatus MeetingStatus `json:"meetingStatus" bson:"meetingStatus"`
	CreatedAt     time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt" bson:"updatedAt"`
}



func CreateMeeting(newMeeting NewMeeting) (Meeting, error) {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    now := time.Now()
    
    // Verify the recipient exists in MongoDB
    var recipient User
    err := database.UsersCollection.FindOne(ctx, bson.M{"_id": newMeeting.Recipient.ID}).Decode(&recipient)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return Meeting{}, errors.New("recipient not found")
        }
        return Meeting{}, err
    }

    // Verify the volunteer exists in MongoDB
    var volunteer User
    err = database.UsersCollection.FindOne(ctx, bson.M{"_id": newMeeting.Volunteer.ID}).Decode(&volunteer)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return Meeting{}, errors.New("volunteer not found")
        }
        return Meeting{}, err
    }

    // Identify services that are available (not already InProgress)
    var availableServices []string
    
    // Create update document for MongoDB
    servicesUpdate := bson.M{"updatedAt": now}
    
    for _, service := range newMeeting.Services {
        if status, exists := recipient.Services[service]; exists && status != InProgress {
            availableServices = append(availableServices, service)
            
            // Update local recipient object's service status
            recipient.Services[service] = InProgress
            
            // Create the field path for this specific service
            serviceFieldPath := fmt.Sprintf("services.%s", service)
            
            // Add to MongoDB update using explicit field path
            servicesUpdate[serviceFieldPath] = string(InProgress)
        }
    }

    // If no services are available, return an error
    if len(availableServices) == 0 {
        return Meeting{}, errors.New("recipient already in progress")
    }

    // Update the recipient's services in MongoDB using explicit field paths
    updateResult, err := database.UsersCollection.UpdateOne(
        ctx,
        bson.M{"_id": recipient.ID},
        bson.M{"$set": servicesUpdate},
    )
    
    if err != nil {
        return Meeting{}, err
    }
    
    if updateResult.ModifiedCount == 0 {
        return Meeting{}, errors.New("failed to update recipient services")
    }
    
    // Log successful update
    fmt.Printf("Updated recipient %s services: %+v\n", recipient.ID, updateResult)

    // Create a new meeting with reference IDs and a new MongoDB ObjectID
    meeting := Meeting{
        ID:            primitive.NewObjectID().Hex(),
        RecipientID:   recipient.ID,
        VolunteerID:   volunteer.ID,
        Date:          newMeeting.Date,
        Services:      availableServices,
        MeetingStatus: newMeeting.MeetingStatus,
        CreatedAt:     now,
        UpdatedAt:     now,
    }

    // For API response, include the full user objects
    meeting.Recipient = recipient
    meeting.Volunteer = volunteer

    // Insert the meeting into MongoDB
    _, err = database.MeetingsCollection.InsertOne(ctx, bson.M{
        "_id":           meeting.ID,
        "recipientId":   meeting.RecipientID,
        "volunteerId":   meeting.VolunteerID,
        "date":          meeting.Date,
        "services":      meeting.Services,
        "meetingStatus": meeting.MeetingStatus,
        "createdAt":     meeting.CreatedAt,
        "updatedAt":     meeting.UpdatedAt,
    })
    if err != nil {
        return Meeting{}, err
    }

    return meeting, nil
}


// if the user is volunteer, update the recipient's service statuses
// if the user is recipient, cancel the meeting, in the client side the recipient will be updated
func CancelMeeting(meetingID string, userUID string) error {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Find the meeting in MongoDB
    var meeting Meeting
    err := database.MeetingsCollection.FindOne(ctx, bson.M{"_id": meetingID}).Decode(&meeting)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return errors.New("meeting not found")
        }
        return err
    }

    // Get the user who is cancelling
    var user User
    err = database.UsersCollection.FindOne(ctx, bson.M{"_id": userUID}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return errors.New("user not found")
        }
        return err
    }

    // If a volunteer is cancelling, update recipient's service statuses
    if user.Role == Volunteer {
        var recipient User
        err = database.UsersCollection.FindOne(ctx, bson.M{"_id": meeting.RecipientID}).Decode(&recipient)
        if err != nil {
            if err == mongo.ErrNoDocuments {
                return errors.New("recipient not found")
            }
            return err
        }

        // Create a map of field updates for MongoDB
        servicesUpdate := make(map[string]interface{})
        for _, service := range meeting.Services {
            servicesUpdate[fmt.Sprintf("services.%s", service)] = string(NeedAssistance)
        }

        // Update the recipient's services in MongoDB
        _, err = database.UsersCollection.UpdateOne(
            ctx,
            bson.M{"_id": recipient.ID},
            bson.M{"$set": servicesUpdate},
        )
        if err != nil {
            return err
        }
    }

    // Delete the meeting from MongoDB
    _, err = database.MeetingsCollection.DeleteOne(ctx, bson.M{"_id": meetingID})
    if err != nil {
        return err
    }

    return nil
}


func GetMeetings(userId string, status MeetingStatus) ([]Meeting, error) {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Build a filter based on the provided parameters
    filter := bson.M{}
    if userId != "" {
        // Filter meetings where user is either recipient or volunteer
        filter["$or"] = []bson.M{
            {"recipientId": userId},
            {"volunteerId": userId},
        }
    }
    if status != "" {
        filter["meetingStatus"] = status
    }

    // Find meetings in MongoDB that match the filter
    cursor, err := database.MeetingsCollection.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    // Decode meetings from MongoDB
    var meetingsData []Meeting
    if err = cursor.All(ctx, &meetingsData); err != nil {
        return nil, err
    }

    // Load user details for each meeting
    var meetings []Meeting
    for _, m := range meetingsData {
        // Get recipient details
        var recipient User
        err = database.UsersCollection.FindOne(ctx, bson.M{"_id": m.RecipientID}).Decode(&recipient)
        if err != nil {
            return nil, fmt.Errorf("failed to load recipient data: %v", err)
        }

        // Get volunteer details
        var volunteer User
        err = database.UsersCollection.FindOne(ctx, bson.M{"_id": m.VolunteerID}).Decode(&volunteer)
        if err != nil {
            return nil, fmt.Errorf("failed to load volunteer data: %v", err)
        }

        // Populate meeting with user details for API response
        m.Recipient = recipient
        m.Volunteer = volunteer
        meetings = append(meetings, m)
    }

    return meetings, nil
}


func UpdateMeetingStatus(meetingID string, newStatus MeetingStatus) (Meeting, error) {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Find the meeting in MongoDB
    var meeting Meeting
    err := database.MeetingsCollection.FindOne(ctx, bson.M{"_id": meetingID}).Decode(&meeting)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return Meeting{}, errors.New("meeting not found")
        }
        return Meeting{}, err
    }

    // Update the meeting status in MongoDB
    now := time.Now()
    _, err = database.MeetingsCollection.UpdateOne(
        ctx,
        bson.M{"_id": meetingID},
        bson.M{"$set": bson.M{
            "meetingStatus": newStatus,
            "updatedAt":     now,
        }},
    )
    if err != nil {
        return Meeting{}, err
    }

    // Update the meeting object with the new status
    meeting.MeetingStatus = newStatus
    meeting.UpdatedAt = now
    
    // Load user details for API response
    var recipient User
    err = database.UsersCollection.FindOne(ctx, bson.M{"_id": meeting.RecipientID}).Decode(&recipient)
    if err != nil {
        return Meeting{}, fmt.Errorf("failed to load recipient data: %v", err)
    }

    var volunteer User
    err = database.UsersCollection.FindOne(ctx, bson.M{"_id": meeting.VolunteerID}).Decode(&volunteer)
    if err != nil {
        return Meeting{}, fmt.Errorf("failed to load volunteer data: %v", err)
    }

    // Set user data for API response
    meeting.Recipient = recipient
    meeting.Volunteer = volunteer

    return meeting, nil
}
