// Code generated by "stringer -type=DiagKind"; DO NOT EDIT.

package diag

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Fatal-0]
	_ = x[Error-1]
	_ = x[Warning-2]
	_ = x[Info-3]
}

const _DiagKind_name = "FatalErrorWarningInfo"

var _DiagKind_index = [...]uint8{0, 5, 10, 17, 21}

func (i DiagKind) String() string {
	if i < 0 || i >= DiagKind(len(_DiagKind_index)-1) {
		return "DiagKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _DiagKind_name[_DiagKind_index[i]:_DiagKind_index[i+1]]
}
