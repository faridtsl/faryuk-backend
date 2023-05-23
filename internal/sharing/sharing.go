package sharing

import (
	"FaRyuk/internal/types"

	"github.com/google/uuid"
)

// NewSharing : returns new sharing using an owner the result and the user
func NewSharing(owner, result, user string) *types.Sharing {
	id := uuid.New().String()

	return &types.Sharing{id, user, owner, result, "Pending"}
}
