package applets

import "fmt"

type AppletFunc func(args []string) error

type Applet struct {
	Name        string
	Description string
	Run         AppletFunc
}

var Registry = map[string]Applet{}

func Register(name string, description string, run AppletFunc) {
	Registry[name] = Applet{
		Name:        name,
		Description: description,
		Run:         run,
	}
}

func GetApplet(name string) (Applet, bool) {
	applet, ok := Registry[name]
	return applet, ok
}

func RunApplet(name string, args []string) error {
	applet, ok := Registry[name]
	if !ok {
		return fmt.Errorf("unknown applet: %s", name)
	}

	return applet.Run(args)
}
