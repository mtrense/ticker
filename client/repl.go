package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"

	"github.com/mtrense/ticker/eventstore"

	"github.com/c-bata/go-prompt"
)

func (s *TickerClient) NewPrompt() *prompt.Prompt {
	return prompt.New(s.execute, s.complete,
		prompt.OptionTitle("Ticker client shell"),
		prompt.OptionPrefix(" » "),
	)
}

func (s *TickerClient) execute(line string) {
	args := strings.Split(line, " ")
	cmd := args[0]
	args = args[1:]
	switch cmd {
	case "state", "s":
		s.replState()
	case "connect", "c":
		s.replConnect()
	case "append", "a":
		s.replAppend()
	}
}

func (s *TickerClient) replState() {
	fmt.Printf("connectedTo: %s\nclientName: %s\n", s.address, s.clientName)
	if s.connection != nil {
		fmt.Printf("Connection established\n")
	} else {
		fmt.Printf("Not connected\n")
	}
}

func (s *TickerClient) replConnect() {
	go s.Connect()
}

func (s *TickerClient) replAppend() {
	types := []string{
	}
	var eventType string
	survey.AskOne(&survey.Select{
		Options: types,
	}, &eventType)
	et := strings.Split(eventType, ":")
	//eventType := strings.Split(prompt.Choose("Type = ", types), ":")
	prompt.Input("Aggregate ID = ", noopCompleter)
	ev := eventstore.Event{
		ID:         "1234",
		Aggregate:  []string{"de", et[0], "12345"},
		Type:       et[1],
		OccurredAt: time.Now(),
		Revision:   1,
		Payload: map[string]interface{}{
			"first_name": "Fred",
			"last_name":  "Flintstone",
		},
	}
	s.SendAppend(ev)
}

func (s *TickerClient) complete(d prompt.Document) []prompt.Suggest {
	//if d.TextBeforeCursor() == "" {
	return suggestions()
	//}
	//return suggestions(
	//	suggest("state", "Print the current state of the client connection"),
	//	suggest("connect", "Connect to a ticket server"),
	//	suggest("append", "Append an event"),
	//)
}

func suggestions(s ...prompt.Suggest) []prompt.Suggest {
	return s
}

func suggest(text, description string) prompt.Suggest {
	return prompt.Suggest{
		Text:        text,
		Description: description,
	}
}

func noopCompleter(d prompt.Document) []prompt.Suggest {
	return suggestions()
}
