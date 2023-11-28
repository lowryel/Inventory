package middleware

import (
	"fmt"
)

// Helper function to parse user ID from string to uint
func ParseUserID(userID string) uint {
	var parsedID uint
	fmt.Sscanf(userID, "%d", &parsedID)
	return parsedID
}