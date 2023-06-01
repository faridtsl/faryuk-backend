package database

import (
	"FaRyuk/api/utils"
	"FaRyuk/config"
	"FaRyuk/database/mongodb/comment"
	"FaRyuk/database/mongodb/user"
	"FaRyuk/internal/types"
	"errors"
)

type Handler interface {
	CommentHandler
	UserHandler
	// TODO add the remaining interfaces
}

type MainDbHandler struct {
	CommentHandler
	UserHandler
}

type CommentHandler interface {
	Create(r *types.Comment, done chan<- error)
	List(listCommentsFilter utils.ListCommentsFilter, comments chan<- *types.CommentsWithErrorType)
	Delete(id string, done chan<- error)
	Update(r *types.Comment, done chan<- error)
	GetByID(id string, result chan<- *types.CommentWithErrorType)
	CloseCommentDBConnection()
}

type UserHandler interface {
	// Create(r *types.User, done chan<- error)
	// List(comments chan<- *types.UsersWithErrorType)
	// Delete(id string, done chan<- error)
	// Update(r *types.User, done chan<- error)
	// GetByID(id string, result chan<- *types.UserWithErrorType)
	CloseUserDBConnection()
}

func CreateDbHandler(cfg *config.AppConfig) (Handler, error) {
	var dbHandler MainDbHandler
	switch cfg.Database.DbType {
	case "mongo":
		dbHandler.UserHandler = user.NewMongoRepository(cfg)
		dbHandler.CommentHandler = comment.NewMongoRepository(cfg)
	default:
		return nil, errors.New("db type not supported")
	}

	return dbHandler, nil
}
