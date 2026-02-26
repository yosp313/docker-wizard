package utils

const (
	CategoryDatabase     = "database"
	CategoryMessageQueue = "message-queue"
	CategoryCache        = "cache"
	CategoryAnalytics    = "analytics"
	CategoryProxy        = "proxy"
)

func CategoryOrder() []string {
	return []string{
		CategoryDatabase,
		CategoryMessageQueue,
		CategoryCache,
		CategoryAnalytics,
		CategoryProxy,
	}
}

func CategoryLabel(category string) string {
	switch category {
	case CategoryDatabase:
		return "Databases"
	case CategoryMessageQueue:
		return "Message Queues"
	case CategoryCache:
		return "Caching"
	case CategoryAnalytics:
		return "Analytics"
	case CategoryProxy:
		return "Webservers / Proxies"
	default:
		return category
	}
}
