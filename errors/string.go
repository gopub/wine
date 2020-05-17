package errors

type errString string

func (e errString) Error() string {
	return string(e)
}

const (
	NotExists errString = "not exists"
)

func String(s string) error {
	return errString(s)
}
