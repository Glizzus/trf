package domain

import "fmt"

// Rating is a type that represents the rating of a claim.
type Rating string

// String returns the string representation of the rating.
func (r Rating) String() string {
	return string(r)
}

// ParseRating parses a string into a Rating.
// If the string is not a valid rating, an error is returned.
func ParseRating(s string) (Rating, error) {
	if _, ok := ratingsOpposite[s]; !ok {
		return "", fmt.Errorf("invalid rating: %s", s)
	}
	return Rating(s), nil
}

// Opposite returns the opposite rating of the current rating.
func (r Rating) Opposite() Rating {
	return Rating(ratingsOpposite[r.String()])
}

// This map serves two purposes:
// 1. The keys are all of the valid ratings - this is used to validate the rating
// 2. The value for each key is the opposite of the key. This is used to generate the opposite rating.
var ratingsOpposite = map[string]string{
	// These are strict opposites
	"True":                "False",
	"Mostly True":         "Mostly False",
	"Mostly False":        "Mostly True",
	"False":               "True",
	"Legit":               "Fake",
	"Fake":                "Legit",
	"Correct Attribution": "Misattributed",
	"Misattributed":       "Correct Attribution",

	// These just get flipped to true because they are vague
	"Unproven":             "True",
	"Unfounded":            "True",
	"Outdated":             "True",
	"Miscaptioned":         "True",
	"Legend":               "True",
	"Scam":                 "True",
	"Labeled Satire":       "True",
	"Originated as Satire": "True",

	// These are neutral, so they stay the same
	"Research in Progress": "Research in Progress",
	"Mixture":              "Mixture",
	"Lost Legend":          "Lost Legend",
	"Recall":               "Recall",
}
