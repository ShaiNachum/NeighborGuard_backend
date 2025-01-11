package schemas

import "neighborguard/pkg/services"

type SearchUsersResponseSchema struct {
	Users []services.User `json:"users"`
}
