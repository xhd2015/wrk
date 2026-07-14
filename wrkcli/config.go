package wrkcli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config is the optional user config at $WRK_HOME/config.json.
type Config struct {
	Version int            `json:"version"`
	Create  *CreateSection `json:"create,omitempty"`
}

// CreateSection holds create-mode UX options.
// Legacy create.interceptor is deliberately omitted so leftover JSON is ignored.
type CreateSection struct {
	Window   *CreateWindow   `json:"window,omitempty"`
	Terminal *CreateTerminal `json:"terminal,omitempty"`
	Agent    *CreateAgent    `json:"agent,omitempty"`
}

// CreateWindow configures Mission Control Desktop creation.
type CreateWindow struct {
	// Mode is "new" when window UX is on; absent/empty means off.
	Mode string `json:"mode,omitempty"`
}

// CreateTerminal configures iTerm2 open mode.
type CreateTerminal struct {
	// Mode is "new" | "reuse" | "smart"; absent/empty means terminal off.
	Mode string `json:"mode,omitempty"`
}

// CreateAgent configures agent-run launch after create.
type CreateAgent struct {
	Enabled        *bool    `json:"enabled,omitempty"`
	Runner         string   `json:"runner,omitempty"`
	PromptTemplate string   `json:"prompt_template,omitempty"`
	Args           []string `json:"args,omitempty"`
}

// loadConfig reads $WRK_HOME/config.json. Missing file returns (nil, nil).
func loadConfig(wrkHome string) (*Config, error) {
	path := filepath.Join(wrkHome, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("wrk: read config.json: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("wrk: parse config.json: %w", err)
	}
	return &cfg, nil
}

// loadConfigMap reads $WRK_HOME/config.json as a generic map so unknown keys
// can be preserved across management writes. Missing file returns (nil, nil).
func loadConfigMap(wrkHome string) (map[string]interface{}, error) {
	path := filepath.Join(wrkHome, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("wrk: read config.json: %w", err)
	}
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("wrk: parse config.json: %w", err)
	}
	if root == nil {
		root = map[string]interface{}{}
	}
	return root, nil
}

// saveConfigMap writes config.json with indent + trailing newline via temp+rename.
func saveConfigMap(wrkHome string, root map[string]interface{}) error {
	if root == nil {
		root = map[string]interface{}{}
	}
	path := filepath.Join(wrkHome, "config.json")
	if err := os.MkdirAll(wrkHome, 0o755); err != nil {
		return fmt.Errorf("wrk: create WRK_HOME: %w", err)
	}
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("wrk: marshal config.json: %w", err)
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("wrk: write config.json: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("wrk: write config.json: %w", err)
	}
	return nil
}

// ensureCreateMap returns the create object under root, creating it if needed.
func ensureCreateMap(root map[string]interface{}) (map[string]interface{}, error) {
	if root == nil {
		return nil, fmt.Errorf("wrk: nil config map")
	}
	createVal, ok := root["create"]
	if !ok || createVal == nil {
		createMap := map[string]interface{}{}
		root["create"] = createMap
		return createMap, nil
	}
	createMap, ok := createVal.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("wrk: config.json create must be an object")
	}
	return createMap, nil
}
