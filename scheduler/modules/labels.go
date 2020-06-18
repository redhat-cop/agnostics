package modules

import (
	"github.com/redhat-gpe/scheduler/config"
	"sort"
	"math/rand"
)

func LabelPredicates(clouds map[string]config.Cloud, labels map[string]string) map[string]config.Cloud {
	result := map[string]config.Cloud{}

out:
	for k, v := range clouds {
		// Check if all labels match
		for lk, lv := range labels {
			if v.Labels[lk] != lv {
				continue out
			}
		}
		// All labels match, we can keep that cloud
		result[k] = v
	}
	return result
}

// Priorities

// ByLabels implements sort.Interface for []config.Cloud
// based on the weight
type ByWeight []config.Cloud
func (a ByWeight) Len() int           { return len(a) }
func (a ByWeight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByWeight) Less(i, j int) bool { return a[i].Weight >= a[j].Weight }

func applyPriorityWeight(clouds map[string]config.Cloud, preferences map[string]string) []config.Cloud {
	result := []config.Cloud{}
	for _, v := range clouds {
		for kp, vp := range preferences {
			if vl, ok := v.Labels[kp]; ok {
				if vl == vp {
					v.Weight = v.Weight + 1
				}
			}
		}
		result = append(result, v)
	}
	return result
}


func LabelPriorities(clouds map[string]config.Cloud, preferences map[string]string) []config.Cloud {
	result := applyPriorityWeight(clouds, preferences)
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	sort.Sort(ByWeight(result))
	return result
}
