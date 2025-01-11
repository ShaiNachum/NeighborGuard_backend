package handlers

import (
	"encoding/json"
	"neighborguard/api/schemas"
	"neighborguard/pkg/services"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GetUsers godoc
// @Summary Get all users
// @Description Get all users with optional filters
// @Tags users
// @Produce json
// @Param email query string false "Email to filter users"
// @Param role query string false "Role to filter users"
// @Param filterByLat query float64 false "Filter by latitude"
// @Param filterByLon query float64 false "Filter by longitude"
// @Param isRequiredAssistance query bool false "Is required assistance"
// @Success 200 {object} schemas.SearchUsersResponseSchema
// @Failure 404 {object} map[string]string{}
// @Router /users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	var filterByLat, filterByLon *float64
	if filterByLatStr := r.URL.Query().Get("filterByLat"); filterByLatStr != "" {
		if val, err := strconv.ParseFloat(filterByLatStr, 64); err == nil {
			filterByLat = &val
		}
	}
	if filterByLonStr := r.URL.Query().Get("filterByLon"); filterByLonStr != "" {
		if val, err := strconv.ParseFloat(filterByLonStr, 64); err == nil {
			filterByLon = &val
		}
	}

	role := services.Role(r.URL.Query().Get("role"))
	isRequiredAssistance := r.URL.Query().Get("isRequiredAssistance") == "true"

	users, err := services.GetUsers(email, filterByLat, filterByLon, role, isRequiredAssistance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := schemas.SearchUsersResponseSchema{Users: users}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetNearbyRecipients godoc
// @Summary Get nearby recipients needing assistance
// @Description Get recipients who need assistance matching volunteer's languages and services
// @Tags users
// @Produce json
// @Param volunteerUID query string true "Volunteer's UID"
// @Param filterByLat query float64 false "Filter by latitude"
// @Param filterByLon query float64 false "Filter by longitude"
// @Success 200 {object} schemas.SearchUsersResponseSchema
// @Failure 404 {object} map[string]string{}
// @Failure 500 {object} map[string]string{}
// @Router /users/recipients [get]
func GetNearbyRecipients(w http.ResponseWriter, r *http.Request) {
	volunteerUID := r.URL.Query().Get("volunteerUID")

	var filterByLat, filterByLon *float64
	if filterByLatStr := r.URL.Query().Get("filterByLat"); filterByLatStr != "" {
		if val, err := strconv.ParseFloat(filterByLatStr, 64); err == nil {
			filterByLat = &val
		}
	}
	if filterByLonStr := r.URL.Query().Get("filterByLon"); filterByLonStr != "" {
		if val, err := strconv.ParseFloat(filterByLonStr, 64); err == nil {
			filterByLon = &val
		}
	}

	recipients, err := services.GetNearbyRecipients(volunteerUID, filterByLat, filterByLon)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := schemas.SearchUsersResponseSchema{Users: recipients}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user
// @Tags user
// @Accept json
// @Produce json
// @Param user body services.NewUser true "User object that needs to be created"
// @Success 200 {object} services.User
// @Failure 400 {object} map[string]string{}
// @Router /user [post]
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var newUser services.NewUser

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := services.CreateUser(newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateUser godoc
// @Summary Update an existing user
// @Description Update an existing user's information
// @Tags user
// @Accept json
// @Produce json
// @Param uid path string true "User ID"
// @Param user body services.User true "Updated user information"
// @Success 200
// @Failure 400 {object} map[string]string{}
// @Failure 404 {object} map[string]string{}
// @Router /users/{uid} [put]
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	var updatedUser services.User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := services.UpdateUser(uid, updatedUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetUserByEmail godoc
// @Summary Get user by email
// @Description Get a single user by their email address
// @Tags user
// @Produce json
// @Param email path string true "User email"
// @Success 200 {object} services.User
// @Failure 404 {object} map[string]string{}
// @Failure 500 {object} map[string]string{}
// @Router /user/{email} [get]
func GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	user, err := services.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
