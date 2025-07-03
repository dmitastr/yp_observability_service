package errs

import "errors"

var ErrorWrongPath error = errors.New("path is not supported")
var ErrorWrongUpdateType error = errors.New("update type is not supported")
var ErrorMetricDoesNotExist error = errors.New("metric was not found")
var ErrorMetricTableEmpty error = errors.New("no metrics added yet")
var ErrorValueFromEmptyMetric = errors.New("getting value from empty metric is not allowed")
