package plans

import "errors"

var (
	ErrPlanNotFound = errors.New("plan not found")
	ErrInvalidPlanID = errors.New("invalid plan ID")
)
