// Code generated by "stringer -type=Location -linecomment"; DO NOT EDIT.

package digiposte

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[LocationInbox-0]
	_ = x[LocationSafe-1]
	_ = x[LocationTrash-2]
}

const _Location_name = "INBOXSAFETRASH"

var _Location_index = [...]uint8{0, 5, 9, 14}

func (i Location) String() string {
	if i < 0 || i >= Location(len(_Location_index)-1) {
		return "Location(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Location_name[_Location_index[i]:_Location_index[i+1]]
}
