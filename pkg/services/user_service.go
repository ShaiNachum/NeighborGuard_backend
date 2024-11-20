package services

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Gender string
type Role string

const (
	Male   Gender = "male"
	Female Gender = "female"
)

const (
	Volunteer Role = "volunteer"
	Recipient Role = "recipient"
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

type Meeting struct {
	Recipient User      `json:"recipient"`
	Volunteer User      `json:"volunteer"`
	Date      time.Time `json:"date"`
}

type NewUser struct {
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Age         int      `json:"age"`
	PhoneNumber int      `json:"phoneNumber"`
	Gender      Gender   `json:"gender"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Address     Address  `json:"address"`
	Languages   []string `json:"languages"`
	Services    []string `json:"services"`
	Role        Role     `json:"role"`
	LonLat      LonLat   `json:"lonLat"`
}

type User struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Age         int       `json:"age"`
	PhoneNumber int       `json:"phoneNumber"`
	Gender      Gender    `json:"gender"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	Address     Address   `json:"address"`
	Languages   []string  `json:"languages"`
	Services    []string  `json:"services"`
	Role        Role      `json:"role"`
	LonLat      LonLat    `json:"lonLat"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ExtendedUser struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Age         int       `json:"age"`
	PhoneNumber int       `json:"phoneNumber"`
	Gender      Gender    `json:"gender"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	Address     Address   `json:"address"`
	Languages   []string  `json:"languages"`
	Services    []string  `json:"services"`
	Role        Role      `json:"role"`
	Meetings    []Meeting `json:"meetings"`
	LonLat      LonLat    `json:"lonLat"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

var (
	usersStore    = make(map[string]User)
	meetingsStore = make(map[string]Meeting)
	mu            sync.Mutex
)

func GetUsers(email string, toExtendMeeting bool) ([]ExtendedUser, error) {
	mu.Lock()
	defer mu.Unlock()

	if len(usersStore) == 0 {
		return nil, errors.New("no users found")
	}

	var users []ExtendedUser
	for _, user := range usersStore {
		if email == "" || user.Email == email {
			var meetings []Meeting
			if toExtendMeeting {
				for _, meeting := range meetingsStore {
					if user.Role == Volunteer && meeting.Volunteer.ID == user.ID {
						meetings = append(meetings, meeting)
					} else if user.Role == Recipient && meeting.Recipient.ID == user.ID {
						meetings = append(meetings, meeting)
					}
				}
			}
			extendedUser := ExtendedUser{
				ID:          user.ID,
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				Age:         user.Age,
				PhoneNumber: user.PhoneNumber,
				Gender:      user.Gender,
				Email:       user.Email,
				Password:    user.Password,
				Address:     user.Address,
				Languages:   user.Languages,
				Services:    user.Services,
				Role:        user.Role,
				Meetings:    meetings,
				LonLat:      user.LonLat,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
			}
			users = append(users, extendedUser)
		}

	}

	return users, nil
}

func CreateUser(newUser NewUser) (User, error) {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	newID := fmt.Sprintf("user_%d", len(usersStore)+1)

	newUserModel := User{
		ID:          newID,
		FirstName:   newUser.FirstName,
		LastName:    newUser.LastName,
		Age:         newUser.Age,
		PhoneNumber: newUser.PhoneNumber,
		Gender:      newUser.Gender,
		Email:       newUser.Email,
		Password:    newUser.Password,
		Address:     newUser.Address,
		Languages:   newUser.Languages,
		Services:    newUser.Services,
		Role:        newUser.Role,
		LonLat:      newUser.LonLat,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	usersStore[newID] = newUserModel
	return newUserModel, nil
}
