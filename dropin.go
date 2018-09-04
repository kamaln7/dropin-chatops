package dropin

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/kamaln7/dropin-chatops/config"
	"github.com/kballard/go-shellquote"
)

type Dropin struct {
	Config *config.Config

	mux *http.ServeMux
}

func New(config *config.Config) *Dropin {
	dropin := &Dropin{
		Config: config,
	}

	return dropin
}

func (d *Dropin) Serve() error {
	d.mux = http.NewServeMux()
	d.mux.HandleFunc("/", d.httpHandler)

	return http.ListenAndServe(d.Config.ListenAddress, d.mux)
}

func (d *Dropin) httpHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err := url.ParseQuery(string(body))
	if err != nil {
		log.Printf("error parsing request body: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	authed := true
	authed = authed && payload.Get("token") == d.Config.Token
	authed = authed && payload.Get("team_id") == d.Config.Team

	if !authed {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	text := payload.Get("text")
	commandParts, err := shellquote.Split(text)
	if err != nil || len(commandParts) < 1 {
		w.Write([]byte("couldn't parse arguments"))
		return
	}

	var (
		commandName = commandParts[0]
		commandArgs = commandParts[1:]
	)

	op := &Operation{
		Channel:     payload.Get("channel_id"),
		User:        payload.Get("user_id"),
		ResponseURL: payload.Get("response_url"),
		Text:        text,
		Command:     d.getCommand(commandName),
		Config:      d.Config,
		Args:        commandArgs,
	}

	go op.Process()
}

func (d *Dropin) getCommand(name string) *config.Command {
	var c *config.Command

	for _, command := range d.Config.Commands {
		if command.Name == name {
			c = command
			break
		}
	}

	return c
}
