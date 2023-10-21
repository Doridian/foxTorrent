package bencoding

import "errors"

var ErrMissingRequiredField = errors.New("missing required field")
var ErrInvalidType = errors.New("invalid type")
