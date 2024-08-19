package extension

func ErrString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
