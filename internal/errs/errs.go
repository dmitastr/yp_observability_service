package errs

import "errors"

var ErrorWrongPath = errors.New("path is not supported")
var ErrorWrongUpdateType = errors.New("update type is not supported")
var ErrorMetricDoesNotExist = errors.New("metric was not found")
var ErrorMetricTableEmpty = errors.New("no metrics added yet")
var ErrorValueFromEmptyMetric = errors.New("getting value from empty metric is not allowed")
