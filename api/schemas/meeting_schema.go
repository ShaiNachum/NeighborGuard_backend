package schemas

import "neighborguard/pkg/services"

type SearchMeetingsResponseSchema struct {
    Meetings []services.Meeting `json:"meetings"`
}