package handlers

import (
	"encoding/json"
	"neighborguard/api/schemas"
	"neighborguard/pkg/services"
	"net/http"
)


// GetUsers godoc
// @Summary Get all users
// @Description Get all users, optionally filtered by email
// @Tags users
// @Produce json
// @Param email query string false "Email to filter users"
// @Param toExtendMeeting query bool false "To extend meeting"
// @Success 200 {object} schemas.SearchUsersResponseSchema
// @Failure 404 {object} map[string]string{}
// @Router /users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	toExtendMeeting := r.URL.Query().Get("toExtendMeeting") == "true"

	users, err := services.GetUsers(email, toExtendMeeting)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    response := schemas.SearchUsersResponseSchema{Users: users}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body services.NewUser true "User object that needs to be created"
// @Success 200 {object} services.User
// @Failure 400 {object} map[string]string{}
// @Router /users [post]
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

func EditUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("EditUser"))
}
