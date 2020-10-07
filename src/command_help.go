package src

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type CommandHelp struct {
	Commands map[string]Command
}

func (this *CommandHelp) Help() string {
	return "Usage: terrarium help [command]"
}

func (this *CommandHelp) Run(config *Config, args []string) error {
	if len(args) == 1 {
		fmt.Println(this.Help())
	}
	if len(args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	if len(args) > 2 {
		return errors.New("Expected no more than one positional argument for 'help' command: <command>")
	}

	command, ok := this.Commands[args[1]]
	if !ok {
		return fmt.Errorf("Invalid command '%s'", args[1])
	}

	fmt.Println(command.Help())

	return nil
}
