package models

// CPU gathers few information about processor
type CPU struct {
	ModelName string `json:"model_name,omitempty"`
	Vendor    string `json:"vendor,omitempty"`
	Cores     int    `json:"cores,omitempty"`
	// Usage     float64 `json:"usage"`
}

type Disk struct {
	Name       string       `json:"name,omitempty"`
	Model      string       `json:"model,omitempty"`
	Size       uint64       `json:"size,omitempty"`
	Type       string       `json:"type,omitempty"`
	Controller string       `json:"controller,omitempty"`
	Partitions []*Partition `json:"partitions,omitempty"`
}

type Partition struct {
	Name     string `json:"name,omitempty"`
	Size     uint64 `json:"size,omitempty"`
	Type     string `json:"type,omitempty"`
	ReadOnly bool   `json:"read_only,omitempty"`
}
