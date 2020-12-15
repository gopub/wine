package httpvalue

const (
	StatusTransportFailed = 600
)

func IsValidStatus(s int) bool {
	return s >= 100 && s <= 999
}
