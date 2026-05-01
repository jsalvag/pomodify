package notifications

import "fyne.io/fyne/v2"

type sender interface {
	SendNotification(notification *fyne.Notification)
}

type Notifier struct {
	app sender
}

func New(app sender) *Notifier {
	return &Notifier{app: app}
}

func (n *Notifier) Send(title, content string) {
	if n == nil || n.app == nil {
		return
	}

	n.app.SendNotification(&fyne.Notification{
		Title:   title,
		Content: content,
	})
}
