package database

import (
	"FaRyuk/api/utils"
	"FaRyuk/config"
	"FaRyuk/database/mongodb/comment"
	"FaRyuk/internal/types"
	"errors"
)

type DbHandler interface {
	CommentHandler
	// TODO add the remaining interfaces
}

type MainDb struct {
	CommentHandler
	// TODO add the remaining handlers
}

type CommentHandler interface {
	Create(r *types.Comment, done chan<- error)
	List(listCommentsFilter utils.ListCommentsFilter, comments chan<- *types.CommentsWithErrorType)
	Delete(id string, done chan<- error)
	Update(r *types.Comment, done chan<- error)
	GetByID(id string, result chan<- *types.CommentWithErrorType)
	// GetCommentsByText(search string, result chan *types.CommentsWithErrorType)
	// GetCommentsByTextAndOwner(search string, idUser string, result chan *types.CommentsWithErrorType)
	// GetCommentsByResultID(idResult string, result chan<- *types.CommentsWithErrorType)
	CloseConnection()
}

func CreateDbHandler(cfg *config.AppConfig) (DbHandler, error) {
	var dbHandler MainDb
	switch cfg.Database.DbType {
	case "mongo":
		dbHandler.CommentHandler = comment.NewMongoRepository(cfg)
	default:
		return nil, errors.New("db type not supported")
	}
	return dbHandler, nil
}
