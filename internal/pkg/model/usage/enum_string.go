// Code generated by "stringer -type=Enum"; DO NOT EDIT.

package usage

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Unknown-0]
	_ = x[Monthly-1]
	_ = x[Daily-2]
}

const _Enum_name = "UnknownMonthlyDaily"

var _Enum_index = [...]uint8{0, 7, 14, 19}

func (i Enum) String() string {
	if i < 0 || i >= Enum(len(_Enum_index)-1) {
		return "Enum(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Enum_name[_Enum_index[i]:_Enum_index[i+1]]
}
