package app

import (
	"github.com/SeeJson/backend_scaffold/cerror"
	"github.com/SeeJson/backend_scaffold/model"
)

func MapErr(err error) error {
	if model.IsNotFoundError(err) {
		return cerror.ErrNotFound
	}
	if model.IsDuplicatedEntryError(err) {
		return cerror.ErrDuplicatedEntry
	}
	if aerr, ok := err.(*cerror.APIError); ok {
		return aerr
	}
	return cerror.ErrInternalServerError
}
