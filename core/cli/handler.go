package cli

import "os"

type Command func() error

var commands = map[string]Command{
	"build": Build,
	"help":  Help,
}

func HandlerCmd() error {
	if len(os.Args) < 1 {
		cmd := commands[os.Args[1]]
		return cmd()

	}
	cmd := commands[os.Args[1]]
	return cmd()
}

func Help() error {
	helpText := `Usage: candy <command> [options]

Commands:
  build   Build the project
  help    Show this help message

Options:
  -h, --help   Show this help message
`
	println(helpText)
	return nil
}
