package spoof

import "context"

// MockSpoofer is a Spoofer that prepends "NOT" to the message.
// This is used for testing purposes.
type MockSpoofer struct{}

// Spoof returns the message prepended with "NOT".
// This will never return an error.
func (m *MockSpoofer) Spoof(ctx context.Context, message, rating string) (string, error) {
	return "NOT " + message, nil
}
