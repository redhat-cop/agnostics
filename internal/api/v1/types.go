package v1

// The Cloud is the main object that the scheduler work with.
type Cloud struct {
	Name string `json:"name"`
	Labels map[string]string `json:"labels"`
	// Weight is usually not provided by config, but automatically filled
	// by the scheduler later, depending on priorities configured.
	// It's possible to add it to the config though, if needed.
	Weight int `json:"weight"`
	// Enabled defines if the cloud can be selected when loading the config. It's a top-level control. If it's set to false, then the cloud will not be used. It takes precedence over scheduling, thus over taints and tolerations.
	// True by default.
	Enabled bool `json:"enabled"`
	// Taints are part of the mechanism to reduce the priority (effect=PreferNoSchedule)
	// or set a cloud as unschedulable (effect=NoSchedule).
	// The taints prevail over the labels or any other control.
	// They can be bypass when requesting schedule to the scheduler by
	// specifying tolerations. In other words, a taint allows a cloud to refuse
	// deployment to be scheduled unless that deployment has a matching toleration.
	// Taints are usually dynamic resources managed by the scheduler, but they can also
	// be statically provided in the configuration.
	Taints map[string]Taint `json:"taints"`
}

type Error struct {
	Code int32 `json:"code"`
	Message string `json:"message"`
}

type Message struct {
	Message string `json:"message"`
}

type ScheduleQuery struct {
	CloudSelector map[string]string `json:"cloud_selector"`
	CloudPreference map[string]string `json:"cloud_preference"`
	Tolerations []Toleration `json:"tolerations"`
	UUID string `json:"uuid,omitempty"`
}

type GitCommit struct {
	Hash string `json:"hash"`
	Author string `json:"author"`
	Date string `json:"date"`
}

const(
	// Do not allow new deployments to be scheduled onto the cloud
	// unless they tolerate the taint
	TaintEffectNoSchedule string = "NoSchedule"
	// Like TaintEffectNoSchedule, but the scheduler tries not to schedule
	// new deployments onto the cloud, rather than prohibiting new deployment
	// from scheduling onto the cloud entirely.
	TaintEffectPreferNoSchedule string = "PreferNoSchedule"

	TolerationOpExists string = "Exists"
	TolerationOpEqual  string = "Equal"
)


// The cloud this taint is attached to has the "effect" on any deployment that
// does not tolerate the Taint.
type Taint struct {
	// Required. The taint key to be applied to a cloud.
	Key string `json:"key"`
	// The taint value corresponding to the taint key.
	// +optional
	Value string `json:"value"`
	// Required. The effect of the taint on deployments that do not tolerate the taint.
	// Valid effects are NoSchedule, PreferNoSchedule.
	Effect string `json:"effect"`
}

// The deployment this Toleration is attached to tolerates any taint that matches
// the triple <key,value,effect> using the matching operator <operator>.
type Toleration struct {
	// Key is the taint key that the toleration applies to. Empty means match all taint keys.
	// If the key is empty, operator must be Exists; this combination means to match all values and all keys.
	// +optional
	Key string `json:"key,omitempty"`
	// Operator represents a key's relationship to the value.
	// Valid operators are Exists and Equal. Defaults to Equal.
	// Exists is equivalent to wildcard for value, so that a cloud can
	// tolerate all taints of a particular category.
	// +optional
	Operator string `json:"operator,omitempty"`
	// Value is the taint value the toleration matches to.
	// If the operator is Exists, the value should be empty, otherwise just a regular string.
	// +optional
	Value string `json:"value,omitempty"`
	// Effect indicates the taint effect to match. Empty means match all taint effects.
	// When specified, allowed values are NoSchedule, PreferNoSchedule.
	// +optional
	Effect string `json:"effect,omitempty"`
}

// Placement object to track where a deployment goes. The link is done thanks to uuid.
type Placement struct {
	// The uuid of the CloudForms service
	UUID string `json:"uuid,omitempty"`
	// Date the placement was made. UTF and RFC3339
	Date string `json:"date"`
	// The cloud where it was scheduled to.
	Cloud Cloud `json:"cloud"`
}
