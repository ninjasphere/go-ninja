package channels

// TODO: MOST OF THIS

type Notification struct {
	Title    string               `json:"title"`
	Subtitle string               `json:"subtitle"`
	Priority notificationPriority `json:"priority"`
	Category notificationCategory `json:"category"`
}

type notificationPriority string

const (
	NotificationPriorityMax     = "max"
	NotificationPriorityHigh    = "high"
	NotificationPriorityDefault = "default"
	NotificationPriorityLow     = "low"
	NotificationPriorityMin     = "min"
)

type notificationCategory string

const (
	NotificationCategoryAlert      = "alert"
	NotificationCategoryQuery      = "query"
	NotificationCategorySuggestion = "suggestion"
)
