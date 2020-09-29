package v1

import (
	"fmt"
	"time"
)


// MatchTaint checks if the taint matches taintToMatch. Taints are unique by key:effect,
// if the two taints have same key:effect, regard as they match.
func (t Taint) MatchTaint(taintToMatch Taint) bool {
	return t.Key == taintToMatch.Key && t.Effect == taintToMatch.Effect
}

// ToString converts taint struct to string in format '<key>=<value>:<effect>', '<key>=<value>:', '<key>:<effect>', or '<key>'.
func (t Taint) String() string {
	if len(t.Effect) == 0 {
		if len(t.Value) == 0 {
			return fmt.Sprintf("%v", t.Key)
		}
		return fmt.Sprintf("%v=%v:", t.Key, t.Value)
	}
	if len(t.Value) == 0 {
		return fmt.Sprintf("%v:%v", t.Key, t.Effect)
	}
	return fmt.Sprintf("%v=%v:%v", t.Key, t.Value, t.Effect)
}

// Taint method taints a cloud with the provided Taint
func (c *Cloud) Taint (t Taint) {
	for i, v := range c.Taints {
		if t.MatchTaint(v) {
			// already tainted with that taint
			c.Taints[i] = t
			return
		}
	}
	c.Taints = append(c.Taints, t)
}

func NewTaint() Taint {
	return Taint{
		CreationTimestamp: time.Now().UTC().Round(time.Second),
	}
}

func (taint Taint) IsTolerated(tolerations []Toleration) bool {
	for _, tol := range tolerations {
		if tol.ToleratesTaint(taint) {
			return true
		}
	}
	return false
}
