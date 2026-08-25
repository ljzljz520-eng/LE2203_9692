package store

import (
	"encoding/json"
	"fmt"
)

func marshal(value any) ([]byte, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode value: %w", err)
	}
	return b, nil
}

func unmarshal(data []byte, target any) error {
	if len(data) == 0 {
		return fmt.Errorf("decode value: empty payload")
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode value: %w", err)
	}
	return nil
}
