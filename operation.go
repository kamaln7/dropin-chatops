package dropin

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/kamaln7/dropin-chatops/config"
)

type Operation struct {
	Channel, User, ResponseURL, Text string
	Command                          *config.Command
	Config                           *config.Config
	Args                             []string
}

func (op *Operation) sendMessage(format string, args ...interface{}) error {
	sr := &SlackResponse{
		Text:         fmt.Sprintf(format, args...),
		ResponseType: "in_channel",
	}

	return sr.Send(op.ResponseURL)
}

func (op *Operation) Process() {
	if !(op.authenticated() && op.commandExists()) {
		return
	}

	err := op.sendMessage("running `%s`", op.Command.Name)
	if err != nil {
		log.Printf("error sending response: %v\n", err)
	}

	op.runCommand()
}

func (op *Operation) runCommand() {
	ctx, cancel := context.WithTimeout(context.Background(), op.Command.TimeoutDuration)
	defer cancel()

	args := make([]string, len(op.Command.Args))
	copy(args, op.Command.Args)
	if op.Command.TakesArguments {
		args = append(args, op.Args...)
	} else {
		if len(op.Args) > 0 {
			serr := op.sendMessage("warning: command `%s` does not accept arguments", op.Command.Name)

			if serr != nil {
				log.Printf("error sending response: %v\n", serr)
			}
		}
	}

	cmd := exec.CommandContext(ctx, op.Command.Executable, args...)
	cmd.Dir = op.Command.Chdir

	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			serr := op.sendMessage("command `%s` exceeded deadline of %s", op.Command.Name, op.Command.TimeoutDuration.String())

			if serr != nil {
				log.Printf("error sending response: %v\n", serr)
			}
		} else {
			serr := op.sendMessage("error running command `%s`:\n```\n%s\n```", op.Command.Name, err)

			if serr != nil {
				log.Printf("error sending response: %v\n", serr)
			}
		}
	}

	output := string(out)
	if output == "" {
		output = "no output"
	}
	err = op.sendMessage("output:\n```\n%s\n```", output)
	if err != nil {
		log.Printf("error sending response: %v\n", err)
	}
}

func (op *Operation) authenticated() bool {
	found := false
	for _, user := range op.Config.Users {
		if user.ID == op.User {
			found = true
			break
		}
	}

	if !found {
		err := op.sendMessage("unauthenticated")
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		return false
	}

	found = false
	for _, channel := range op.Config.Channels {
		if channel.ID == op.Channel {
			found = true
			break
		}
	}

	if !found {
		err := op.sendMessage("unauthenticated")
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
	}

	return found
}

func (op *Operation) commandExists() bool {
	if op.Command != nil {
		return true
	}

	var commandList []string
	for _, command := range op.Config.Commands {
		commandList = append(commandList, command.Name)
	}

	op.sendMessage("command `%s` not found. available commands: %s", op.Text, strings.Join(commandList, ", "))
	return false
}
