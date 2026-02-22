package truenas

import "strings"

// isNotFoundError checks if an API error indicates a resource was not found.
// TrueNAS returns errors containing "does not exist" or "[ENOENT]" for missing resources.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "[ENOENT]") ||
		strings.Contains(msg, "not found") ||
		strings.Contains(msg, "no such instance")
}
