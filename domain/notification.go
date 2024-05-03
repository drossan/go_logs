package domain

// Notifier define el puerto para enviar notificaciones.
type Notifier interface {
	SendNotification(message string) error
}
