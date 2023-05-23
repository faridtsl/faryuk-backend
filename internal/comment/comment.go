package comment

import (
	"time"

	"FaRyuk/internal/types"

	"github.com/google/uuid"
)

// NewComment : constructs a comment from content, owner and result
func NewComment(content string, owner string, result string) *types.Comment {
	id := uuid.New().String()

	return &types.Comment{
		ID:          id,
		Content:     content,
		IDResult:    result,
		Owner:       owner,
		CreatedDate: time.Now(),
		UpdatedDate: time.Now(),
	}
}
