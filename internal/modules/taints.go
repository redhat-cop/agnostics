package modules

import (
	"github.com/redhat-gpe/agnostics/internal/api/v1"
	"github.com/redhat-gpe/agnostics/internal/log"
	"sort"
)

// TaintPredicates filters out clouds with taints unless there are matching tolerations.
func TaintPredicates(clouds []v1.Cloud, tolerations []v1.Toleration) []v1.Cloud {
	result := []v1.Cloud{}

	for _, cloud := range clouds {
		if cloud.IsTolerated(tolerations, v1.TaintEffectNoSchedule) {
			result = append(result, cloud)
		}
	}
	return result
}


// TaintPriorities sort the clouds according to the Taints present, and the weight passed.
func TaintPriorities(clouds []v1.Cloud, tolerations []v1.Toleration, weight int) []v1.Cloud {
	for i, c := range clouds {
		for _, t := range c.Taints {
			// For priorities, we're only intereste in PreferNoSchedule
			if t.Effect == v1.TaintEffectPreferNoSchedule {
				if ! t.IsTolerated(tolerations) {
					clouds[i].Weight = c.Weight - weight
				} else {
					log.Debug.Println(clouds[i].Name, "taint", t, "tolerated")
				}
			}
		}
	}
	sort.Sort(ByWeight(clouds))
	return clouds
}
