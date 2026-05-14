package vector

import (
	"encoding/json"
	"os"
	"strconv"
)

// LoadNorm reads normalization.json.
func LoadNorm(path string) (*Norm, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	norm := &Norm{}
	if err := json.Unmarshal(raw, norm); err != nil {
		return nil, err
	}
	return norm, nil
}

// LoadMccRisk reads mcc_risk.json into a MccRisk map.
func LoadMccRisk(path string) (*MccRisk, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parsed := map[string]float64{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	var out MccRisk
	for i := range out {
		out[i] = DefaultMccRisk
	}
	for code, risk := range parsed {
		idx, err := strconv.Atoi(code)
		if err == nil && idx >= 0 && idx < len(out) {
			out[idx] = risk
		}
	}
	return &out, nil
}
