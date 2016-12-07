package src

// EngineConfig represents engine object
type EngineConfig struct {
	Name          string
	Owner         string
	Version       float32
	Product       string
	Arch          string
	EngineConfig  map[string]string
	Hypervisors   []map[string]string
	Storages      []map[string]string
	ExtraStorages []map[string]string
}
