package applets

import "fmt"

type Applet struct {
	Name string
	Help string
	Exec func(args []string) error
}

var Registry = map[string]Applet{}

func Register(name string, help string, exec func(args []string) error) {
	Registry[name] = Applet{
		Name: name,
		Help: help,
		Exec: exec,
	}
}

func RunApplet(name string, args []string) error {
	applet, ok := Registry[name]
	if !ok {
		return fmt.Errorf("unknown command: %s", name)
	}

	return applet.Exec(args)
}
