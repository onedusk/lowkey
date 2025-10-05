package cobra

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Command struct {
	Use   string
	Short string

	Run  func(cmd *Command, args []string)
	RunE func(cmd *Command, args []string) error

	subCommands []*Command
	parent      *Command
	args        []string
}

func (c *Command) AddCommand(cmds ...*Command) {
	for _, child := range cmds {
		if child == nil {
			continue
		}
		child.parent = c
		c.subCommands = append(c.subCommands, child)
	}
}

func (c *Command) SetArgs(args []string) {
	c.args = append([]string(nil), args...)
}

func (c *Command) Execute() error {
	args := c.args
	if args == nil {
		args = os.Args[1:]
	}
	return c.execute(args)
}

func (c *Command) execute(args []string) error {
	if len(args) == 0 {
		return c.invoke(args)
	}

	next := args[0]
	for _, child := range c.subCommands {
		name := strings.Split(child.Use, " ")[0]
		if name == next {
			return child.execute(args[1:])
		}
	}

	if err := c.invoke(args); err != nil {
		return err
	}
	return fmt.Errorf("unknown command: %s", next)
}

func (c *Command) invoke(args []string) error {
	if c.RunE != nil {
		if err := c.RunE(c, args); err != nil {
			return err
		}
		return nil
	}
	if c.Run != nil {
		c.Run(c, args)
		return nil
	}
	if len(c.subCommands) > 0 {
		return errors.New("no sub-command specified")
	}
	return nil
}

var initializers []func()

func OnInitialize(fns ...func()) {
	initializers = append(initializers, fns...)
}

func ExecuteInitializers() {
	for _, fn := range initializers {
		if fn != nil {
			fn()
		}
	}
}
