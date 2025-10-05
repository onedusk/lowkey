package viper

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Viper struct {
	mu         sync.RWMutex
	configFile string
	store      map[string]string
}

func New() *Viper {
	return &Viper{store: make(map[string]string)}
}

var defaultViper = New()

func (v *Viper) SetConfigFile(path string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.configFile = path
}

func (v *Viper) Set(key string, value interface{}) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.store[key] = toString(value)
}

func (v *Viper) GetString(key string) string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.store[key]
}

func (v *Viper) AutomaticEnv() {}

func (v *Viper) ReadInConfig() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.configFile == "" {
		return errors.New("no config file specified")
	}

	data, err := os.ReadFile(v.configFile)
	if err != nil {
		return err
	}

	if err := v.parse(data); err != nil {
		return err
	}
	return nil
}

func (v *Viper) parse(data []byte) error {
	var m map[string]string
	if err := json.Unmarshal(data, &m); err == nil {
		for k, val := range m {
			v.store[k] = val
		}
		return nil
	}

	lines := bytesSplit(data)
	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		key, val, ok := splitOnce(line, '=')
		if !ok {
			continue
		}
		v.store[string(trimSpace(key))] = string(trimSpace(val))
	}
	return nil
}

func (v *Viper) SafeWriteConfig() error {
	v.mu.RLock()
	path := v.configFile
	v.mu.RUnlock()
	if path == "" {
		return errors.New("no config file specified")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return os.ErrExist
	}
	return os.WriteFile(path, []byte("{}"), 0o600)
}

func SetConfigFile(path string) { defaultViper.SetConfigFile(path) }

func ReadInConfig() error { return defaultViper.ReadInConfig() }

func AutomaticEnv() { defaultViper.AutomaticEnv() }

func Set(key string, value interface{}) { defaultViper.Set(key, value) }

func GetString(key string) string { return defaultViper.GetString(key) }

func NewWithConfig(path string) (*Viper, error) {
	vp := New()
	vp.SetConfigFile(path)
	if err := vp.ReadInConfig(); err != nil {
		return nil, err
	}
	return vp, nil
}

func toString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	case fmt.Stringer:
		return s.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func bytesSplit(data []byte) [][]byte {
	var out [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' || b == '\r' {
			if start < i {
				out = append(out, data[start:i])
			}
			start = i + 1
		}
	}
	if start < len(data) {
		out = append(out, data[start:])
	}
	return out
}

func splitOnce(data []byte, sep byte) ([]byte, []byte, bool) {
	for i, b := range data {
		if b == sep {
			return data[:i], data[i+1:], true
		}
	}
	return nil, nil, false
}

func trimSpace(data []byte) []byte {
	start := 0
	for start < len(data) && isSpace(data[start]) {
		start++
	}
	end := len(data)
	for end > start && isSpace(data[end-1]) {
		end--
	}
	return data[start:end]
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t'
}
