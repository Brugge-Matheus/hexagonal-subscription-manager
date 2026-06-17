package notification

import (
	"fmt"
	"io"
)

type LogNotification struct {
	Out io.Writer
}

func (l LogNotification) Notify(customerID, event, message string) error {
	_, err := fmt.Fprintf(l.Out, "[notification] customer=%s event=%s message=%q\n", customerID, event, message)
	return err
}
