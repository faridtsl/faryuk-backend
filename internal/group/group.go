package group

import (
	"FaRyuk/internal/types"

	"github.com/google/uuid"
)

// NewComment : constructs a comment from content, owner and result
func NewGroup(name string) *types.Group {
	id := uuid.New().String()

	return &types.Group{
		ID:   id,
		Name: name,
	}
}

// ToIDsArray : return an array of IDs from array of interfaces
func ToIDsArray(groups []types.Group) []string {
	res := make([]string, 0)
	for _, v := range groups {
		res = append(res, v.ID)
	}
	return res
}
