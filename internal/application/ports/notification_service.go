package ports

type NotificationService interface {
	Notify(customerID, event, message string) error
}
