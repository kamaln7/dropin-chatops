package config

import (
	"time"

	"github.com/micro/go-config"
	"github.com/micro/go-config/source/file"
)

type User struct {
	Name, ID string
}

type Channel struct {
	Name, ID string
	Users    []string
}

type Command struct {
	Name, Description, Executable, Chdir, Timeout string
	Args                                          []string
	TakesArguments                                bool

	TimeoutDuration time.Duration
}

type Config struct {
	Team, Token, ListenAddress string
	Users                      []*User
	Channels                   []*Channel
	Commands                   []*Command
	Timeout                    string
}

func Read(path string) (*Config, error) {
	conf := config.NewConfig()
	err := conf.Load(file.NewSource(
		file.WithPath(path),
	))
	if err != nil {
		return nil, err
	}

	var c Config
	err = conf.Scan(&c)
	if err != nil {
		return nil, err
	}

	for _, command := range c.Commands {
		var timeout string
		if command.Timeout != "" {
			timeout = command.Timeout
		} else {
			timeout = c.Timeout
		}

		timeoutDuration, err := time.ParseDuration(timeout)
		if err != nil {
			return nil, err
		}

		command.TimeoutDuration = timeoutDuration
	}

	return &c, nil
}
