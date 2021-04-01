package client

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gopkg.in/workanator/go-ataman.v1"

	es "github.com/mtrense/ticker/eventstream/base"
)

type Formatter func(w io.Writer, e *es.Event) error

func OmitPayload(f Formatter) Formatter {
	return func(w io.Writer, e *es.Event) error {
		e.Payload = nil
		return f(w, e)
	}
}

func JsonFormatter(pretty bool) Formatter {
	return func(w io.Writer, e *es.Event) error {
		enc := json.NewEncoder(w)
		if pretty {
			enc.SetIndent("", "  ")
		} else {
			enc.SetIndent("", "")
		}
		return enc.Encode(e)
	}
}

func TextFormatter(pretty bool) Formatter {
	return func(w io.Writer, e *es.Event) error {
		if pretty {
			r := ataman.NewRenderer(ataman.BasicStyle())
			//tpl := "<light+green>%s<->, <bg+light+yellow,black,bold> <<%s>> <-><red>!"
			fmt.Fprint(w, r.MustRenderf("<b>%d<-> » <light+green>%s/%s<->", e.Sequence, strings.Join(e.Aggregate, "."), e.Type))
			if e.Payload != nil {
				fmt.Fprint(w, " »")
				for key, value := range e.Payload {
					fmt.Fprint(w, r.MustRenderf(" <white>%s:<-><blue>%v<->", key, value))
				}
			}
		} else {
			fmt.Fprintf(w, "%d » %s/%s", e.Sequence, strings.Join(e.Aggregate, "."), e.Type)
			if e.Payload != nil {
				fmt.Fprintf(w, " »")
				for key, value := range e.Payload {
					fmt.Fprintf(w, " %s:%v", key, value)
				}
			}
		}
		fmt.Fprintln(w)
		return nil
	}
}
