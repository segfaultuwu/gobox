package applets

func init() {
	Register("sh", "simple shell", Sh)
}

func Sh(args []string) error {
	Sh(args)
	return nil
}
