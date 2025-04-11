package smtp

import (
	"fmt"

	"github.com/hantdev/mitras/consumers/notifiers"
	"github.com/hantdev/mitras/internal/email"
	"github.com/hantdev/mitras/pkg/messaging"
)

const (
	footer          = "Sent by Mitras SMTP Notification"
	contentTemplate = "A publisher with an id %s sent the message over %s with the following values \n %s"
)

var _ notifiers.Notifier = (*notifier)(nil)

type notifier struct {
	agent *email.Agent
}

// New instantiates SMTP message notifier.
func New(agent *email.Agent) notifiers.Notifier {
	return &notifier{agent: agent}
}

func (n *notifier) Notify(from string, to []string, msg *messaging.Message) error {
	subject := fmt.Sprintf(`Notification for Channel %s`, msg.GetChannel())
	if msg.GetSubtopic() != "" {
		subject = fmt.Sprintf("%s and subtopic %s", subject, msg.GetSubtopic())
	}

	values := string(msg.GetPayload())
	content := fmt.Sprintf(contentTemplate, msg.GetPublisher(), msg.GetProtocol(), values)

	return n.agent.Send(to, from, subject, "", "", content, footer)
}
