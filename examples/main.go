package main

import (
	"log"
	"os"

	"github.com/rprtr258/bubbletea/examples/altscreen_toggle"
	"github.com/rprtr258/bubbletea/examples/cellbuffer"
	"github.com/rprtr258/bubbletea/examples/chat"
	"github.com/rprtr258/bubbletea/examples/composable_views"
	"github.com/rprtr258/bubbletea/examples/credit_card_form"
	"github.com/rprtr258/bubbletea/examples/debounce"
	"github.com/rprtr258/bubbletea/examples/exec"
	"github.com/rprtr258/bubbletea/examples/file_picker"
	"github.com/rprtr258/bubbletea/examples/fullscreen"
	"github.com/rprtr258/bubbletea/examples/glamour"
	"github.com/rprtr258/bubbletea/examples/help"
	"github.com/rprtr258/bubbletea/examples/http"
	"github.com/rprtr258/bubbletea/examples/list_default"
	"github.com/rprtr258/bubbletea/examples/list_fancy"
	"github.com/rprtr258/bubbletea/examples/list_simple"
	"github.com/rprtr258/bubbletea/examples/mouse"
	"github.com/rprtr258/bubbletea/examples/package_manager"
	"github.com/rprtr258/bubbletea/examples/pager"
	"github.com/rprtr258/bubbletea/examples/paginator"
	"github.com/rprtr258/bubbletea/examples/pipe"
	"github.com/rprtr258/bubbletea/examples/prevent_quit"
	"github.com/rprtr258/bubbletea/examples/progress_animated"
	"github.com/rprtr258/bubbletea/examples/progress_download"
	"github.com/rprtr258/bubbletea/examples/progress_static"
	"github.com/rprtr258/bubbletea/examples/realtime"
	"github.com/rprtr258/bubbletea/examples/result"
	"github.com/rprtr258/bubbletea/examples/send_msg"
	"github.com/rprtr258/bubbletea/examples/sequence"
	"github.com/rprtr258/bubbletea/examples/simple"
	"github.com/rprtr258/bubbletea/examples/spinner"
	"github.com/rprtr258/bubbletea/examples/spinners"
	"github.com/rprtr258/bubbletea/examples/split_editors"
	"github.com/rprtr258/bubbletea/examples/stopwatch"
	"github.com/rprtr258/bubbletea/examples/table"
	"github.com/rprtr258/bubbletea/examples/tabs"
	"github.com/rprtr258/bubbletea/examples/textarea"
	"github.com/rprtr258/bubbletea/examples/textinput"
	"github.com/rprtr258/bubbletea/examples/textinputs"
	"github.com/rprtr258/bubbletea/examples/timer"
	"github.com/rprtr258/bubbletea/examples/tui_daemon_combo"
	"github.com/rprtr258/bubbletea/examples/views"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run main.go <example>")
	}

	switch os.Args[1] {
	case "--help":
		log.Println("TODO: print examples list")
	case "altscreen-toggle":
		altscreen_toggle.Main()
	case "cellbuffer":
		cellbuffer.Main()
	case "chat":
		chat.Main()
	case "composable-views":
		composable_views.Main()
	case "credit-card-form":
		credit_card_form.Main()
	case "debounce":
		debounce.Main()
	case "exec":
		exec.Main()
	case "file-picker":
		file_picker.Main()
	case "fullscreen":
		fullscreen.Main()
	case "glamour":
		glamour.Main()
	case "help":
		help.Main()
	case "http":
		http.Main()
	case "list-default":
		list_default.Main()
	case "list-fancy":
		list_fancy.Main()
	case "list-simple":
		list_simple.Main()
	case "mouse":
		mouse.Main()
	case "package-manager":
		package_manager.Main()
	case "pager":
		pager.Main()
	case "paginator":
		paginator.Main()
	case "pipe":
		pipe.Main()
	case "prevent-quit":
		prevent_quit.Main()
	case "progress-animated":
		progress_animated.Main()
	case "progress-download":
		progress_download.Main()
	case "progress-static":
		progress_static.Main()
	case "realtime":
		realtime.Main()
	case "result":
		result.Main()
	case "send-msg":
		send_msg.Main()
	case "sequence":
		sequence.Main()
	case "simple":
		simple.Main()
	case "spinner":
		spinner.Main()
	case "spinners":
		spinners.Main()
	case "split-editors":
		split_editors.Main()
	case "stopwatch":
		stopwatch.Main()
	case "table":
		table.Main()
	case "tabs":
		tabs.Main()
	case "textarea":
		textarea.Main()
	case "textinput":
		textinput.Main()
	case "textinputs":
		textinputs.Main()
	case "timer":
		timer.Main()
	case "tui-daemon-combo":
		tui_daemon_combo.Main()
	case "views":
		views.Main()
	default:
		log.Fatal("Unknown example")
	}
}
