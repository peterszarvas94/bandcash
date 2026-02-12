package utils

import (
	"encoding/json"
	"strconv"
)

func ParseRawInt64(value json.RawMessage) (int64, error) {
	if len(value) == 0 {
		return 0, nil
	}

	var asString string
	if err := json.Unmarshal(value, &asString); err == nil {
		if asString == "" {
			return 0, nil
		}
		return strconv.ParseInt(asString, 10, 64)
	}

	var asNumber int64
	if err := json.Unmarshal(value, &asNumber); err == nil {
		return asNumber, nil
	}

	var asFloat float64
	if err := json.Unmarshal(value, &asFloat); err == nil {
		return int64(asFloat), nil
	}

	return 0, strconv.ErrSyntax
}

func ParseRawFloat64(value json.RawMessage) (float64, error) {
	if len(value) == 0 || string(value) == "\"\"" {
		return 0, nil
	}

	var asFloat float64
	if err := json.Unmarshal(value, &asFloat); err == nil {
		return asFloat, nil
	}

	var asString string
	if err := json.Unmarshal(value, &asString); err == nil {
		if asString == "" {
			return 0, nil
		}
		return strconv.ParseFloat(asString, 64)
	}

	return 0, strconv.ErrSyntax
}
