package modelparam

import "github.com/QuantumNous/new-api/common"

const OptionKey = "ModelParameterRules"

func init() {
	common.RegisterOptionUpdateHook(func(key, value string) {
		if key != OptionKey || value == "" {
			return
		}
		if err := ConfigureDefault([]byte(value)); err != nil {
			common.SysError("failed to reload model parameter rules: " + err.Error())
		}
	})
}

type Config struct {
	Version int    `json:"version"`
	Rules   []Rule `json:"rules"`
}

func LoadConfig(data []byte) (*Snapshot, error) {
	var config Config
	if err := common.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return Compile(config.Rules)
}

func ValidateConfig(data string) error {
	_, err := LoadConfig([]byte(data))
	return err
}

// ConfigureDefault validates and atomically publishes a new process-wide rule set.
// A failed configuration leaves the currently active snapshot untouched.
func ConfigureDefault(data []byte) error {
	snapshot, err := LoadConfig(data)
	if err != nil {
		return err
	}
	registry := DefaultRegistry()
	if registry == nil {
		SetDefaultRegistry(NewRegistry(snapshot))
		return nil
	}
	registry.Replace(snapshot)
	return nil
}

func ConfigureFromOptionMap() error {
	common.OptionMapRWMutex.RLock()
	data := common.OptionMap[OptionKey]
	common.OptionMapRWMutex.RUnlock()
	if data == "" {
		return nil
	}
	return ConfigureDefault([]byte(data))
}
