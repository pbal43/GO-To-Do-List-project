package taskerrors

import "errors"

var (
	ErrFoundNothing       = errors.New("found nothing")
	ErrEmptyString        = errors.New("empty inserted string")
	ErrWrongStatus        = errors.New("wrong status")
	ErrTaskIsAlreadyExist = errors.New("task is already exist")
	ErrDBOnGet            = errors.New("db error on get")
	ErrDBOnUpdate         = errors.New("db error on update")
)
