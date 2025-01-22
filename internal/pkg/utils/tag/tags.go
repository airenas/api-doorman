package tag

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	_inPreffix      = "in["
	_betweenPreffix = "between["
)

func Parse(tag string) (string /*key*/, string /*value*/, error) {
	if idx := strings.IndexByte(tag, ':'); idx >= 0 {
		return strings.ToLower(strings.TrimSpace(tag[:idx])), strings.TrimSpace(tag[idx+1:]), nil
	}
	if (strings.TrimSpace(tag)) == "" {
		return "", "", nil
	}
	return "", "", fmt.Errorf("wrong tag value, no ':' in '%s'", tag)
}

func ValidateValue(valueCondition, value string) error {
	if valueCondition == "" {
		return nil
	}
	if value == "" {
		return nil
	}
	if strings.HasPrefix(valueCondition, _inPreffix) && strings.HasSuffix(valueCondition, "]") {
		return validateIn(valueCondition[len(_inPreffix):len(valueCondition)-1], value)
	}
	if strings.HasPrefix(valueCondition, _betweenPreffix) && strings.HasSuffix(valueCondition, "]") {
		return validateBetween(valueCondition[len(_betweenPreffix):len(valueCondition)-1], value)
	}
	return fmt.Errorf("wrong condition: %s", valueCondition)
}

func validateBetween(s, value string) error {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return fmt.Errorf("wrong between condition: %s", s)
	}
	return validateBetweenAsInts(parts[0], parts[1], value)
}

func validateBetweenAsInts(sFrom, sTo, sValue string) error {
	from, err := strconv.Atoi(strings.TrimSpace(sFrom))
	if err != nil {
		return fmt.Errorf("wrong from value: %s: %w", sFrom, err)
	}
	to, err := strconv.Atoi(strings.TrimSpace(sTo))
	if err != nil {
		return fmt.Errorf("wrong to value: %s: %w", sTo, err)
	}
	v, err := strconv.Atoi(strings.TrimSpace(sValue))
	if err != nil {
		return fmt.Errorf("wrong value: %s: %w", sValue, err)
	}
	if v < from || v > to {
		return fmt.Errorf("value '%d' not in range [%d, %d]", v, from, to)
	}
	return nil
}

func validateIn(values, value string) error {
	for _, v := range strings.Split(values, ",") {
		if strings.TrimSpace(v) == value {
			return nil
		}
	}
	return fmt.Errorf("value '%s' not in [%s]", value, values)
}
