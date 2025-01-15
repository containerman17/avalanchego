package valuemeta

const (
	Tombstone byte = 1 << iota // 0b00000001
	Remote                     // 0b00000010
)

const NoFlags byte = 0

func IsTombstone(value []byte) bool {
	if value[0]&Tombstone == 0 {
		return false
	}
	if len(value) != 1 {
		panic("implementation error: Tombstone flag must have length 1")
	}
	if value[0] != Tombstone && value[0]&^Tombstone != 0 {
		panic("implementation error: unexpected bits set in Tombstone flag")
	}
	return true
}

func IsRemote(value []byte) bool {
	return value[0]&Remote != 0
}
