package services

import (
	"errors"
	"fmt"
	"time"
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
	MeetingStatus MeetingStatus `json:"meetingStatus"`
}

type Meeting struct {
	ID            string        `json:"uid"`
	Recipient     User          `json:"recipient"`
	Volunteer     User          `json:"volunteer"`
	Date          int64         `json:"date"`
	MeetingStatus MeetingStatus `json:"meetingStatus"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

func CreateMeeting(newMeeting NewMeeting) (Meeting, error) {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	newID := fmt.Sprintf("meeting_%d", len(meetingsStore)+1)

	// Validate users exist
	recipient, exists := usersStore[newMeeting.Recipient.ID]
	if !exists {
		return Meeting{}, errors.New("recipient not found")
	}
	volunteer, exists := usersStore[newMeeting.Volunteer.ID]
	if !exists {
		return Meeting{}, errors.New("volunteer not found")
	}

	// Add this validation check
	if recipient.AssistanceStatus == InProgress {
		return Meeting{}, errors.New("recipient already in progress")
	}

	// If we get here, the recipient is available, so update their status
	recipient.AssistanceStatus = InProgress
	usersStore[recipient.ID] = recipient

	meeting := Meeting{
		ID:            newID,
		Recipient:     recipient,
		Volunteer:     volunteer,
		Date:          newMeeting.Date,
		MeetingStatus: newMeeting.MeetingStatus,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	meetingsStore[newID] = meeting
	return meeting, nil
}

func CancelMeeting(meetingID string) error {
	mu.Lock()
	defer mu.Unlock()

	// Find the meeting
	meeting, exists := meetingsStore[meetingID]
	if !exists {
		return errors.New("meeting not found")
	}

	// Get the recipient
	recipient, exists := usersStore[meeting.Recipient.ID]
	if !exists {
		return errors.New("recipient not found")
	}

	// Update recipient's status back to NEED_ASSISTANCE
	recipient.AssistanceStatus = NeedAssistance
	usersStore[recipient.ID] = recipient

	// Remove the meeting
	delete(meetingsStore, meetingID)

	return nil
}

func GetMeetings(userId string, status MeetingStatus) ([]Meeting, error) {
    mu.Lock()
    defer mu.Unlock()

    var filteredMeetings []Meeting

    // If no filters provided, return all meetings
    if userId == "" && status == "" {
        // Convert map to slice
        for _, meeting := range meetingsStore {
            filteredMeetings = append(filteredMeetings, meeting)
        }
        return filteredMeetings, nil
    }

    // Filter meetings based on provided parameters
    for _, meeting := range meetingsStore {
        // Check user ID filter if provided
        if userId != "" {
            if meeting.Recipient.ID != userId && meeting.Volunteer.ID != userId {
                continue // Skip if neither recipient nor volunteer matches
            }
        }

        // Check status filter if provided
        if status != "" {
            if meeting.MeetingStatus != status {
                continue // Skip if status doesn't match
            }
        }

        // If we get here, the meeting matches all provided filters
        filteredMeetings = append(filteredMeetings, meeting)
    }

    return filteredMeetings, nil
}


func UpdateMeetingStatus(meetingID string, newStatus MeetingStatus) (Meeting, error) {
    mu.Lock()
    defer mu.Unlock()

    // Find the meeting
    meeting, exists := meetingsStore[meetingID]
    if !exists {
        return Meeting{}, errors.New("meeting not found")
    }

    // Update the status
    meeting.MeetingStatus = newStatus
    meeting.UpdatedAt = time.Now()

    // Save back to store
    meetingsStore[meetingID] = meeting

    return meeting, nil
}
