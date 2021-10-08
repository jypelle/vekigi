package version

import "strconv"

type Version struct {
	MajorNumber int64
	MinorNumber int64
	PatchNumber int64
}

// String generate a human readable Version
func (m *Version) String() string {
	return strconv.FormatInt(m.MajorNumber, 10) + "." + strconv.FormatInt(m.MinorNumber, 10) + "." + strconv.FormatInt(m.PatchNumber, 10)
}

var (
	AppVersion = Version{
		MajorNumber: 1,
		MinorNumber: 0,
		PatchNumber: 0,
	}
)
