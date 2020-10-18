package ert

import (
	"errors"
	"fmt"

	"github.com/f9a/mail"
)

type MailReporter struct {
	*mail.Tx
	from        string
	contentType string
}

func (r MailReporter) To(to ...string) Reporter {
	return func(trace Trace, topic, body string) error {
		if topic != "" {
			topic = fmt.Sprintf("ERT Report: %s: %s", trace, topic)
		} else {
			topic = fmt.Sprintf("ERT Report: %s", trace.String())
		}

		return r.Send(r.from, mail.To(to), mail.Message{
			Topic:       topic,
			Body:        body,
			ContentType: r.contentType,
		})
	}
}

func NewMailReporter(tx *mail.Tx, from, contentType string) (MailReporter, error) {
	if from == "" {
		return MailReporter{}, errors.New("from is required")
	}

	if contentType == "" {
		contentType = "text/plain"
	}

	return MailReporter{
		Tx:          tx,
		from:        from,
		contentType: contentType,
	}, nil
}
