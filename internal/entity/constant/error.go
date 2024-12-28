package constant

import "github.com/pkg/errors"

var ErrorNotFound = errors.New("not found")

// database error
var ErrorNoAffectedRow = errors.New("no affected row")
var ErrorRowNotFound = errors.New("row not found")
