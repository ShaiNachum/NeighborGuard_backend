package handlers

import (
	"encoding/json"
	"neighborguard/api/schemas"
	"neighborguard/pkg/services"
	"net/http"

	"github.com/gorilla/mux"
)

// CreateMeeting godoc
// @Summary Create a new meeting
// @Description Create a new meeting between a volunteer and a recipient
// @Tags meeting
// @Accept json
// @Produce json
// @Param meeting body services.NewMeeting true "Meeting to create"
// @Success 200 {object} services.Meeting
// @Failure 400 {object} map[string]string{}
// @Failure 500 {object} map[string]string{}
// @Router /meeting [post]
func CreateMeeting(w http.ResponseWriter, r *http.Request) {
	var newMeeting services.NewMeeting
	if err := json.NewDecoder(r.Body).Decode(&newMeeting); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	meeting, err := services.CreateMeeting(newMeeting)
	if err != nil {
		// Change this part to handle specific errors
		if err.Error() == "recipient already in progress" {
			http.Error(w, err.Error(), http.StatusConflict) // Use 409 Conflict
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meeting)
}

// CancelMeeting godoc
// @Summary Cancel an existing meeting
// @Description Cancel a meeting and update recipient's status
// @Tags meeting
// @Produce json
// @Param id path string true "Meeting ID"
// @Success 200
// @Failure 404 {object} map[string]string{}
// @Failure 500 {object} map[string]string{}
// @Router /meeting/{id} [delete]
func CancelMeeting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	meetingID := vars["id"]

	err := services.CancelMeeting(meetingID)
	if err != nil {
		if err.Error() == "meeting not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetMeetings godoc
// @Summary Get meetings based on filters
// @Description Get meetings filtered by user ID (recipient or volunteer) and meeting status
// @Tags meetings
// @Produce json
// @Param userId query string false "User ID to filter meetings (can be recipient or volunteer)"
// @Param status query string false "Meeting status to filter (IS_PICKED or DONE)"
// @Success 200 {array} services.Meeting
// @Failure 400 {object} map[string]string{}
// @Failure 500 {object} map[string]string{}
// @Router /meetings [get]
func GetMeetings(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	userId := r.URL.Query().Get("userId")
	status := services.MeetingStatus(r.URL.Query().Get("status"))

	// Validate status if provided
	if status != "" && status != services.IsPicked && status != services.Done {
		http.Error(w, "Invalid meeting status", http.StatusBadRequest)
		return
	}

	meetings, err := services.GetMeetings(userId, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := schemas.SearchMeetingsResponseSchema{Meetings: meetings}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateMeetingStatus godoc
// @Summary Update meeting status
// @Description Update the status of an existing meeting
// @Tags meeting
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Param status body string true "New meeting status (IS_PICKED or DONE)"
// @Success 200 {object} services.Meeting
// @Failure 400 {object} map[string]string{}
// @Failure 404 {object} map[string]string{}
// @Failure 500 {object} map[string]string{}
// @Router /meeting/{id}/status [put]
func UpdateMeetingStatus(w http.ResponseWriter, r *http.Request) {
	// Get meeting ID from URL parameters
	vars := mux.Vars(r)
	meetingID := vars["id"]

	// Get status from query parameter
	status := services.MeetingStatus(r.URL.Query().Get("status"))

	// Validate the status
	if status != services.IsPicked && status != services.Done {
		http.Error(w, "Invalid meeting status", http.StatusBadRequest)
		return
	}

	// Update the meeting status
	updatedMeeting, err := services.UpdateMeetingStatus(meetingID, status)
	if err != nil {
		if err.Error() == "meeting not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedMeeting)
}
