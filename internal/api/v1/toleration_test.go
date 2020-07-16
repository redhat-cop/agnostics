package v1

import (
	"testing"
)

func TestTolerationToleratesTaint(t *testing.T) {
	testCases := []struct {
		description string
		toleration Toleration
		taint Taint
		expected bool
	}{
		{
			description: "Taint and Toleration tolerates with Operator Exists",
			toleration: Toleration{
				Key: "memory-pressure",
				Value: "ignored",
				Effect: TaintEffectNoSchedule,
				Operator: TolerationOpExists,
			},
			taint: Taint{
				Key: "memory-pressure",
				Value: "high",
				Effect: TaintEffectNoSchedule,
			},
			expected: true,
		},
		{
			description: "Taint and Toleration tolerates with Operator Equal",
			toleration: Toleration{
				Key: "memory-pressure",
				Value: "high",
				Effect: TaintEffectNoSchedule,
				Operator: TolerationOpEqual,
			},
			taint: Taint{
				Key: "memory-pressure",
				Value: "high",
				Effect: TaintEffectNoSchedule,
			},
			expected: true,
		},
		{
			description: "Taint and Toleration don't tolerates with Operator Equal (Values are different)",
			toleration: Toleration{
				Key: "memory-pressure",
				Value: "high",
				Effect: TaintEffectNoSchedule,
				Operator: TolerationOpEqual,
			},
			taint: Taint{
				Key: "memory-pressure",
				Value: "critical",
				Effect: TaintEffectNoSchedule,
			},
			expected: false,
		},
		{
			description: "Taint and Toleration don't tolerates with Operator Equal (Keys are different)",
			toleration: Toleration{
				Key: "memory-pressure",
				Value: "high",
				Effect: TaintEffectNoSchedule,
				Operator: TolerationOpEqual,
			},
			taint: Taint{
				Key: "cpu-pressure",
				Value: "high",
				Effect: TaintEffectNoSchedule,
			},
			expected: false,
		},
		{
			description: "Taint and Toleration don't tolerates with invalid Operator",
			toleration: Toleration{
				Key: "memory-pressure",
				Value: "high",
				Effect: TaintEffectNoSchedule,
				Operator: "invalid",
			},
			taint: Taint{
				Key: "memory-pressure",
				Value: "high",
				Effect: TaintEffectNoSchedule,
			},
			expected: false,
		},
	}

	for _, c := range testCases {
		r := c.toleration.ToleratesTaint(c.taint)

		if r != c.expected {
                        t.Errorf("'%s', Expected ToleratesTaint() to be %v but it was %v", c.description, c.expected, r)
		}
        }
}
