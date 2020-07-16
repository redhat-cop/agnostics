package v1

import (
	"strings"
)

// ToleratesTaint checks if the toleration tolerates the taint.
// The matching follows the rules below:
// (1) Empty toleration.effect means to match all taint effects,
//     otherwise taint effect must equal to toleration.effect.
// (2) If toleration.operator is 'Exists', it means to match all taint values.
// (3) Empty toleration.key means to match all taint keys.
//     If toleration.key is empty, toleration.operator must be 'Exists';
//     this combination means to match all taint values and all taint keys.
func (tol Toleration) ToleratesTaint(taint Taint) bool {
	if len(tol.Effect) > 0 && tol.Effect != taint.Effect {
		return false
	}

	if len(tol.Key) > 0 && tol.Key != taint.Key {
		return false
	}

	switch strings.ToLower(tol.Operator) {
	case strings.ToLower(TolerationOpExists):
		return true
	case "", strings.ToLower(TolerationOpEqual):
		return tol.Value == taint.Value
	default:
		return false
	}
}
