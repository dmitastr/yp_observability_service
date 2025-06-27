package errs

import "errors"

var ErrorWrongPath error = errors.New("path is not supported")
var ErrorWrongUpdateType error = errors.New("update type is not supported")
