package modules

import (
	"testing"
	"github.com/redhat-gpe/agnostics/api/v1"
	"sort"
)

// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestTaintPredicates (t *testing.T) {
	clouds := map[string]v1.Cloud{
		"openstack-1": v1.Cloud{
			Name: "openstack-1",
			Taints: map[string]v1.Taint{},
		},
		"openstack-2": v1.Cloud{
			Name: "openstack-2",
			Taints: map[string]v1.Taint{
				"memory-pressure": {
					Key: "memory-pressure",
					Value: "high",
					Effect: v1.TaintEffectPreferNoSchedule,
				},
			},
		},
		"openstack-3": v1.Cloud{
			Name: "openstack-3",
			Taints: map[string]v1.Taint{
				"memory-pressure": {
					Key: "memory-pressure",
					Value: "critical",
					Effect: v1.TaintEffectNoSchedule,
				},
			},
		},
		"openstack-4": v1.Cloud{
			Name: "openstack-4",
			Taints: map[string]v1.Taint{
				"custom-taint": {
					Key: "custom-taint",
					Value: "custom-value",
					Effect: v1.TaintEffectNoSchedule,
				},
			},
		},
		"openstack-5": v1.Cloud{
			Name: "openstack-5",
			Taints: map[string]v1.Taint{
				"cpu-pressure": {
					Key: "cpu-pressure",
					Value: "critical",
					Effect: v1.TaintEffectNoSchedule,
				},
			},
		},
		"openstack-6": v1.Cloud{
			Name: "openstack-6",
			Taints: map[string]v1.Taint{
				"disk-pressure": {
					Key: "disk-pressure",
					Value: "high",
					Effect: v1.TaintEffectNoSchedule,
				},
			},
		},
		"openstack-7": v1.Cloud{
			Name: "openstack-7",
			Taints: map[string]v1.Taint{
				"disk-pressure": {
					Key: "disk-pressure",
					Value: "critical",
					Effect: v1.TaintEffectNoSchedule,
				},
			},
		},
		"openstack-8": v1.Cloud{
			Name: "openstack-8",
			Taints: map[string]v1.Taint{
				"disk-pressure": {
					Key: "disk-pressure",
					Value: "critical",
					Effect: v1.TaintEffectNoSchedule,
				},
				"memory-pressure": {
					Key: "memory-pressure",
					Value: "critical",
					Effect: v1.TaintEffectNoSchedule,
				},
			},
		},
	}
	testCases := []struct {
		description string
		clouds map[string]v1.Cloud
		tolerations []v1.Toleration
		expected []string
	}{
		{
			description: "No tolerations",
			clouds: clouds,
			tolerations: []v1.Toleration{},
			expected: []string{
				"openstack-1",
				"openstack-2",
			},
		},
		{
			description: "Toleration memory-pressure exists",
			clouds: clouds,
			tolerations: []v1.Toleration{
				{
					Key: "memory-pressure",
					Operator: "Exists",
				},
			},
			expected: []string{
				"openstack-1",
				"openstack-2",
				"openstack-3",
			},
		},
		{
			description: "",
			clouds: clouds,
			tolerations: []v1.Toleration{
				{
					Key: "disk-pressure",
					Value: "critical",
					Operator: "Equal",
				},
			},
			expected: []string{
				"openstack-1",
				"openstack-2",
				"openstack-7",
			},
		},
		{
			description: "",
			clouds: clouds,
			tolerations: []v1.Toleration{
				{
					Operator: "Exists",
				},
			},
			expected: []string{
				"openstack-1",
				"openstack-2",
				"openstack-3",
				"openstack-4",
				"openstack-5",
				"openstack-6",
				"openstack-7",
				"openstack-8",
			},
		},

	}

	for _, c := range testCases {
		rclouds := TaintPredicates(c.clouds, c.tolerations)

		r := []string{}
		for _, v := range rclouds {
			r = append(r, v.Name)
		}
		sort.Strings(r)

		if !sliceEqual(r, c.expected) {
                        t.Errorf("'%s', Expected TaintPredicates() to be %v but it was %v", c.description, c.expected, r)
		}
	}
}
