package errors

type errString string

func (e errString) Error() string {
	return string(e)
}

const (
	NotExist errString = "not exist"
)

func String(s string) error {
	return errString(s)
}
