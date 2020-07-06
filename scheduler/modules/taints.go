package modules

import (
	"github.com/redhat-gpe/scheduler/api/v1"
)

type taintedCloud v1.Cloud

func (c taintedCloud) isTolerated(tolerations []v1.Toleration) bool {
taintLoop:
	for _, taint := range c.Taints {
		// Ignore Taint if effect is not NoSchedule
		if taint.Effect != v1.TaintEffectNoSchedule {
			continue taintLoop
		}
	tolerationLoop:
		for _, tol := range tolerations {
			if tol.ToleratesTaint(taint) {
				continue taintLoop
			} else {
				continue tolerationLoop
			}
		}

		return false
	}
	return true
}

// This function filters out clouds with taints unless there are matching tolerations.
func TaintPredicates(clouds map[string]v1.Cloud, tolerations []v1.Toleration) map[string]v1.Cloud {
	result := map[string]v1.Cloud{}

	for k, v := range clouds {
		if (taintedCloud)(v).isTolerated(tolerations) {
			result[k] = v
		}
	}
	return result
}
