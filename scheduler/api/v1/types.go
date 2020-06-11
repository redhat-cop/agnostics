package v1

type Error struct {
	Code int32 `json:"code"`
	Message string `json:"message"`
}

type Message struct {
	Message string `json:"message"`
}

type CloudQuery struct {
	CloudSelector map[string]string `json:"cloud_selector"`
	CloudPreference map[string]string `json:"cloud_preference"`
}
