package services

import (
	"sync"
)

var (
	usersStore    = make(map[string]User)
	meetingsStore = make(map[string]Meeting)
	mu            sync.Mutex
)
