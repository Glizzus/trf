package spoofing

import "context"

// Spoofer is an interface for generating spoofed messages.
// The Spoofer is used to generate a spoofed message based on a given message and rating.
type Spoofer interface {
	Spoof(ctx context.Context, message, rating string) (string, error)
}
