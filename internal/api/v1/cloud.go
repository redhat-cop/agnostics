package v1

// NewCloud is the constructor for a new Cloud object.
func NewCloud() Cloud {
	return Cloud{Enabled: true}
}

func (c Cloud) IsTolerated(tolerations []Toleration, effect string) bool {
taintLoop:
	for _, taint := range c.Taints {
		// Ignore Taint if effect is not the one passed
		if taint.Effect != effect {
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
