package modules

import (
	"github.com/redhat-gpe/agnostics/internal/api/v1"
	"sort"
	"math/rand"
)

func LabelPredicates(clouds []v1.Cloud, labels map[string]string) []v1.Cloud {
	result := []v1.Cloud{}

out:
	for _, v := range clouds {
		if v.Enabled == false {
			continue out
		}
		// Check if all labels match
		for lk, lv := range labels {
			if v.Labels[lk] != lv {
				continue out
			}
		}
		// All labels match, we can keep that cloud
		result = append(result, v)
	}
	return result
}

// Priorities

// ByLabels implements sort.Interface for []v1.Cloud
// based on the weight
type ByWeight []v1.Cloud
func (a ByWeight) Len() int           { return len(a) }
func (a ByWeight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByWeight) Less(i, j int) bool { return a[i].Weight >= a[j].Weight }

func applyPriorityWeight(clouds []v1.Cloud, preferences map[string]string, weight int) []v1.Cloud {
	result := []v1.Cloud{}
	for _, v := range clouds {
		for kp, vp := range preferences {
			if vl, ok := v.Labels[kp]; ok {
				if vl == vp {
					v.Weight = v.Weight + weight
				}
			}
		}
		result = append(result, v)
	}
	return result
}


func LabelPriorities(clouds []v1.Cloud, preferences map[string]string, weight int) []v1.Cloud {
	result := applyPriorityWeight(clouds, preferences, weight)
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	sort.Sort(ByWeight(result))
	return result
}
