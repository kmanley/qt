package parser

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/therecipe/qt/internal/utils"
)

const (
	SIGNAL = "signal"
	SLOT   = "slot"

	IMPURE = "impure"
	PURE   = "pure"

	PLAIN            = "plain"
	CONSTRUCTOR      = "constructor"
	COPY_CONSTRUCTOR = "copy-constructor"
	MOVE_CONSTRUCTOR = "move-constructor"
	DESTRUCTOR       = "destructor"

	CONNECT    = "Connect"
	DISCONNECT = "Disconnect"
	CALLBACK   = "callback"

	GETTER = "getter"
	SETTER = "setter"

	VOID = "void"

	TILDE = "~"

	MOC = "moc"
)

func IsPackedList(v string) bool {
	return (strings.HasPrefix(v, "QList<") ||
		strings.HasPrefix(v, "QVector<") ||
		strings.HasPrefix(v, "QStack<") ||
		strings.HasPrefix(v, "QQueue<")) &&

		strings.Count(v, "<") == 1 &&
		!strings.Contains(v, ":") &&
		State.ClassMap[UnpackedList(v)] != nil
}

func CleanValue(v string) string {
	for _, b := range []string{"*", "const", "&amp", "&", ";"} {
		v = strings.Replace(v, b, "", -1)
	}
	return strings.TrimSpace(v)
}

func CleanName(name, value string) string {
	switch name {
	case
		"type",
		"func",
		"range",
		"string",
		"int",
		"map",
		"const",
		"interface",
		"select",
		"strings",
		"new",
		"signal",
		"ptr",
		"register":
		{
			return name[:len(name)-2]
		}

	case "":
		{
			return fmt.Sprintf("v%v", strings.Replace(strings.ToLower(CleanValue(value)[:2]), ".", "", -1))
		}
	}

	return name
}

func UnpackedList(v string) string {
	return CleanValue(strings.Split(strings.Split(v, "<")[1], ">")[0])
}

var LibDeps = map[string][]string{
	"Core":          {"Widgets", "Gui"}, //Widgets, Gui
	"AndroidExtras": {"Core"},
	"Gui":           {"Widgets", "Core"}, //Widgets
	"Network":       {"Core"},
	"Xml":           {"XmlPatterns", "Core"}, //XmlPatterns
	"DBus":          {"Core"},
	"Nfc":           {"Core"},
	"Script":        {"Core"},
	"Sensors":       {"Core"},
	"Positioning":   {"Core"},
	"Widgets":       {"Gui", "Core"},
	"Sql":           {"Widgets", "Gui", "Core"}, //Widgets, Gui
	"MacExtras":     {"Gui", "Core"},
	"Qml":           {"Network", "Core"},
	"WebSockets":    {"Network", "Core"},
	"XmlPatterns":   {"Network", "Core"},
	"Bluetooth":     {"Core"},
	"WebChannel":    {"Network", "Qml", "Core"}, //Network (needed for static linking ios)
	"Svg":           {"Widgets", "Gui", "Core"},
	"Multimedia":    {"MultimediaWidgets", "Widgets", "Network", "Gui", "Core"},   //MultimediaWidgets, Widgets
	"Quick":         {"QuickWidgets", "Widgets", "Network", "Qml", "Gui", "Core"}, //QuickWidgets, Widgets, Network (needed for static linking ios)
	"Help":          {"Sql", "CLucene", "Network", "Widgets", "Gui", "Core"},      //Sql + CLucene + Network (needed for static linking ios)
	"Location":      {"Positioning", "Quick", "Gui", "Core"},
	"ScriptTools":   {"Script", "Widgets", "Core"}, //Script, Widgets
	"UiTools":       {"Widgets", "Gui", "Core"},
	"X11Extras":     {"Gui", "Core"},
	"WinExtras":     {"Gui", "Core"},
	"WebEngine":     {"Widgets", "WebEngineWidgets", "WebChannel", "Network", "WebEngineCore", "Quick", "Gui", "Qml", "Core"}, //Widgets, WebEngineWidgets, WebChannel, Network
	"TestLib":       {"Widgets", "Gui", "Core"},                                                                               //Widgets, Gui
	"SerialPort":    {"Core"},
	"SerialBus":     {"Core"},
	"PrintSupport":  {"Widgets", "Gui", "Core"},
	//"PlatformHeaders": []string{}, //TODO: uncomment
	"Designer": {"UiPlugin", "Widgets", "Gui", "Xml", "Core"},
	"Scxml":    {"Network", "Qml", "Core"}, //Network (needed for static linking ios)
	"Gamepad":  {"Gui", "Core"},

	"Purchasing":        {"Core"},
	"DataVisualization": {"Gui", "Core"},
	"Charts":            {"Widgets", "Gui", "Core"},
	//"Quick2DRenderer":   {}, //TODO: uncomment

	//"NetworkAuth":    {"Network", "Gui", "Core"},
	"Speech":         {"Core"},
	"QuickControls2": {"Core"},

	"Sailfish": {"Core"},

	MOC:         make([]string, 0),
	"build_ios": {"Core", "Gui", "Network", "Sql", "Xml", "DBus", "Nfc", "Script", "Sensors", "Positioning", "Widgets", "Qml", "WebSockets", "XmlPatterns", "Bluetooth", "WebChannel", "Svg", "Multimedia", "Quick", "Help", "Location", "ScriptTools", "MultimediaWidgets", "UiTools", "PrintSupport"},
}

var Libs = []string{
	"Core",
	"AndroidExtras",
	"Gui",
	"Network",
	"Xml",
	"DBus",
	"Nfc",
	"Script", //depreached (planned) in 5.6
	"Sensors",
	"Positioning",
	"Widgets",
	"Sql",
	"MacExtras",
	"Qml",
	"WebSockets",
	"XmlPatterns",
	"Bluetooth",
	"WebChannel",
	"Svg",
	"Multimedia",
	"Quick",
	"Help",
	"Location",
	"ScriptTools", //depreached (planned) in 5.6
	"UiTools",
	"X11Extras",
	"WinExtras",
	"WebEngine",
	"TestLib",
	"SerialPort",
	"SerialBus",
	"PrintSupport",
	//"PlatformHeaders", //missing imports/guards
	"Designer",
	"Scxml",
	"Gamepad",

	"Purchasing",        //GPLv3 & LGPLv3
	"DataVisualization", //GPLv3
	"Charts",            //GPLv3
	//"Quick2DRenderer",   //GPLv3

	//"NetworkAuth",
	"Speech",
	"QuickControls2",

	"Sailfish",
}

func ShouldBuild(module string) bool {
	return true
}

func GetLibs() []string {
	var out = Libs
	for i := len(Libs) - 1; i >= 0; i-- {
		switch {
		case !(runtime.GOOS == "darwin" || runtime.GOOS == "linux") && Libs[i] == "WebEngine",
			runtime.GOOS != "windows" && Libs[i] == "WinExtras",
			runtime.GOOS != "darwin" && Libs[i] == "MacExtras",
			runtime.GOOS != "linux" && Libs[i] == "X11Extras":
			{
				out = append(out[:i], out[i+1:]...)
			}

		case utils.QT_VERSION() != "5.8.0" && Libs[i] == "Speech":
			{
				out = append(out[:i], out[i+1:]...)
			}
		}
	}
	return out
}

func Dump() {
	for _, c := range State.ClassMap {
		var bb = new(bytes.Buffer)
		defer bb.Reset()

		fmt.Fprint(bb, "funcs\n\n")
		for _, f := range c.Functions {
			fmt.Fprintln(bb, f)
		}

		fmt.Fprint(bb, "\n\nenums\n\n")
		for _, e := range c.Enums {
			fmt.Fprintln(bb, e)
		}

		utils.MkdirAll(utils.GoQtPkgPath("internal", "binding", "dump", c.Module))
		utils.SaveBytes(utils.GoQtPkgPath("internal", "binding", "dump", c.Module, fmt.Sprintf("%v.txt", c.Name)), bb.Bytes())
	}
}

func SortedClassNamesForModule(module string, template bool) []string {
	var output = make([]string, 0)
	for _, class := range State.ClassMap {
		if class.Module == module {
			output = append(output, class.Name)
		}
	}
	sort.Stable(sort.StringSlice(output))

	if State.Moc && template {
		var items = make(map[string]string)

		for _, cn := range output {
			if class, exist := State.ClassMap[cn]; exist {
				items[cn] = class.Bases
			}
		}

		var tmpOutput = make([]string, 0)

		for len(items) > 0 {
			for item, dep := range items {

				var c, exist = State.ClassMap[dep]
				if exist && c.Module != MOC {
					tmpOutput = append(tmpOutput, item)
					delete(items, item)
					continue
				}

				for _, key := range tmpOutput {
					if key == dep {
						tmpOutput = append(tmpOutput, item)
						delete(items, item)
						break
					}
				}

			}
		}
		output = tmpOutput
	}

	return output
}

func SortedClassesForModule(module string, template bool) []*Class {
	var (
		classNames = SortedClassNamesForModule(module, template)
		output     = make([]*Class, len(classNames))
	)
	for i, name := range classNames {
		output[i] = State.ClassMap[name]
	}
	return output
}
