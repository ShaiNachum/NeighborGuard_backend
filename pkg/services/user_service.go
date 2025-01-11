package services

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/umahmood/haversine"
)

type Gender string
type Role string
type AssistanceStatus string

const (
	Male   Gender = "MALE"
	Female Gender = "FEMALE"
)

const (
	Volunteer Role = "VOLUNTEER"
	Recipient Role = "RECIPIENT"
)

const (
	DoNotNeedAssistance AssistanceStatus = "DO_NOT_NEED_ASSISTANCE"
	NeedAssistance      AssistanceStatus = "NEED_ASSISTANCE"
	InProgress          AssistanceStatus = "IN_PROGRESS"
)


type LonLat struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type Address struct {
	City            string `json:"city"`
	Street          string `json:"street"`
	HouseNumber     int    `json:"houseNumber"`
	ApartmentNumber int    `json:"apartmentNumber"`
}


type NewUser struct {
	FirstName        string           `json:"firstName"`
	LastName         string           `json:"lastName"`
	Age              int              `json:"age"`
	PhoneNumber      string           `json:"phoneNumber"`
	Gender           Gender           `json:"gender"`
	Email            string           `json:"email"`
	Password         string           `json:"password"`
	Address          Address          `json:"address"`
	Languages        []string         `json:"languages"`
	Services         []string         `json:"services"`
	Role             Role             `json:"role"`
	LonLat           LonLat           `json:"lonLat"`
	LastOK           int64            `json:"lastOK"`
	ProfileImage     string           `json:"profileImage"`
	AssistanceStatus AssistanceStatus `json:"assistanceStatus"`
}

type User struct {
	ID               string           `json:"uid"`
	FirstName        string           `json:"firstName"`
	LastName         string           `json:"lastName"`
	Age              int              `json:"age"`
	PhoneNumber      string           `json:"phoneNumber"`
	Gender           Gender           `json:"gender"`
	Email            string           `json:"email"`
	Password         string           `json:"password"`
	Address          Address          `json:"address"`
	Languages        []string         `json:"languages"`
	Services         []string         `json:"services"`
	Role             Role             `json:"role"`
	LonLat           LonLat           `json:"lonLat"`
	LastOK           int64            `json:"lastOK"`
	ProfileImage     string           `json:"profileImage"`
	AssistanceStatus AssistanceStatus `json:"assistanceStatus"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}


func GetNearbyRecipients(
	volunteerUID string,
	filterByLat *float64,
	filterByLon *float64,
) ([]User, error) {
	mu.Lock()
	defer mu.Unlock()

	// Get and validate volunteer
	volunteer, exists := usersStore[volunteerUID]
	if !exists {
		return nil, errors.New("volunteer not found")
	}
	if volunteer.Role != Volunteer {
		return nil, errors.New("only volunteers can use this endpoint")
	}

	var recipients []User
	// Find matching recipients
	for _, user := range usersStore {
		// Skip non-recipients
		if user.Role != Recipient {
			continue
		}

		if user.AssistanceStatus == InProgress {
			continue
		}

		// Check location if filters provided
		if filterByLat != nil && filterByLon != nil {
			nearLocation := LonLat{Longitude: *filterByLon, Latitude: *filterByLat}
			if !isInLocation(user.LonLat, nearLocation) {
				continue
			}
		}

		// Check if user needs assistance
		if !checkIfUserNeedsAssistance(user) {
			continue
		}

		// Check if they share any languages
		if !hasCommonElements(volunteer.Languages, user.Languages) {
			continue
		}

		// Check if they share any services
		if len(user.Services) > 0 {
			if !hasCommonElements(volunteer.Services, user.Services) {
				continue
			}
		}

		recipients = append(recipients, user)
	}

	// Sort recipients by priority
	sort.Slice(recipients, func(i, j int) bool {
		// First priority: LastOK time (older = higher priority)
		if recipients[i].LastOK != recipients[j].LastOK {
			return recipients[i].LastOK < recipients[j].LastOK
		}
		// Second priority: Number of matching services with volunteer
		matchingServicesI := countMatchingServices(volunteer.Services, recipients[i].Services)
		matchingServicesJ := countMatchingServices(volunteer.Services, recipients[j].Services)
		return matchingServicesI > matchingServicesJ
	})

	return recipients, nil
}

func GetUsers(
	email string,
	filterByLat *float64,
	filterByLon *float64,
	role Role,
	isRequiredAssistance bool,
) ([]User, error) {
	mu.Lock()
	defer mu.Unlock()

	if len(usersStore) == 0 {
		return nil, errors.New("no users found")
	}

	var users []User
	for _, user := range usersStore {
		if email != "" && user.Email != email { //if i pass an email it returns only the user with that email otherwise it returns all users
			continue
		}

		if filterByLat != nil && filterByLon != nil {
			nearLocation := LonLat{Longitude: *filterByLon, Latitude: *filterByLat}

			if !isInLocation(user.LonLat, nearLocation) {
				continue
			}
		}

		if role != "" && role != user.Role {
			continue
		}

		if isRequiredAssistance && (user.Role == Volunteer || user.Role == Recipient && !checkIfUserNeedsAssistance(user)) {
			continue
		}

		users = append(users, user)
	}

	return users, nil
}

func CreateUser(newUser NewUser) (User, error) {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	newID := fmt.Sprintf("user_%d", len(usersStore)+1)

	newUserModel := User{
		ID:               newID,
		FirstName:        newUser.FirstName,
		LastName:         newUser.LastName,
		Age:              newUser.Age,
		PhoneNumber:      newUser.PhoneNumber,
		Gender:           newUser.Gender,
		Email:            newUser.Email,
		Password:         newUser.Password,
		Address:          newUser.Address,
		Languages:        newUser.Languages,
		Services:         newUser.Services,
		Role:             newUser.Role,
		LonLat:           newUser.LonLat,
		LastOK:           now.Unix(),
		ProfileImage:     newUser.ProfileImage,
		AssistanceStatus: newUser.AssistanceStatus,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	usersStore[newID] = newUserModel
	return newUserModel, nil
}

func UpdateUser(uid string, updatedUser User) error {
	mu.Lock()
	defer mu.Unlock()

	// Check if user exists
	_, exists := usersStore[uid]
	if !exists {
		return errors.New("user not found")
	}

	// Update user data
	user := usersStore[uid]
	user.FirstName = updatedUser.FirstName
	user.LastName = updatedUser.LastName
	user.PhoneNumber = updatedUser.PhoneNumber
	user.Languages = updatedUser.Languages
	user.Services = updatedUser.Services
	user.Address = updatedUser.Address
	user.LonLat = updatedUser.LonLat
	user.LastOK = updatedUser.LastOK
	user.ProfileImage = updatedUser.ProfileImage
	user.AssistanceStatus = updatedUser.AssistanceStatus
	user.UpdatedAt = time.Now()

	// Save back to store
	usersStore[uid] = user

	return nil
}

func GetUserByEmail(email string) (*User, error) {
	mu.Lock()
	defer mu.Unlock()

	for _, user := range usersStore {
		if user.Email == email {
			return &user, nil
		}
	}

	return nil, errors.New("user not found")
}

// Helper functions:

// Helper function to count matching services
func countMatchingServices(volunteerServices, recipientServices []string) int {
	count := 0
	serviceSet := make(map[string]bool)
	for _, service := range volunteerServices {
		serviceSet[service] = true
	}
	for _, service := range recipientServices {
		if serviceSet[service] {
			count++
		}
	}
	return count
}

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

// Helper function to check if the user needs assistance
func checkIfUserNeedsAssistance(user User) bool {
	// Check if user needs assistance based on LastOK time
	timeBasedNeed := time.Now().Unix()-user.LastOK > 60 // 1 minute for testing (should be 24*60*60 for 24 hours)

	// Check if recipient has services
	hasServices := len(user.Services) > 0

	// User needs assistance if either condition is met
	needsAssistance := timeBasedNeed || hasServices

	// Update status if user needs assistance and isn't already marked
	if needsAssistance && user.AssistanceStatus != NeedAssistance {
		user.AssistanceStatus = NeedAssistance
		usersStore[user.ID] = user
	}

	return needsAssistance
}
