package enums

import (
	"fmt"
)

type Status string

const (
	InReview    Status = "inreview"
	Approved    Status = "approved"
	Rejected    Status = "rejected"
	Paused      Status = "paused"
	Disabled    Status = "disabled"
	Unsubmitted Status = "unsubmitted"
)

type StatusMap map[string]Status

var (
	statusMap StatusMap = StatusMap{
		"inreview":    InReview,
		"approved":    Approved,
		"rejected":    Rejected,
		"paused":      Paused,
		"disabled":    Disabled,
		"unsubmitted": Unsubmitted,

		//aliases
		"received": InReview,
	}
)

func StatusFromString(strStatus string) (Status, error) {
	status, exists := statusMap[strStatus]
	if !exists {
		return "", fmt.Errorf("invalid status string: %s", strStatus)
	}
	return status, nil
}
