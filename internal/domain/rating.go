package domain

import "fmt"

type Rating string

func (r Rating) String() string {
	return string(r)
}

func ParseRating(s string) (Rating, error) {
	if _, ok := ratingsOpposite[s]; !ok {
		return "", fmt.Errorf("invalid rating: %s", s)
	}
	return Rating(s), nil
}

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
	"Labeled Satire":         "True",
	"Originated as Satire": "True",

	// These are neutral, so they stay the same
	"Research in Progress": "Research in Progress",
	"Mixture":              "Mixture",
	"Lost Legend":          "Lost Legend",
	"Recall":               "Recall",
}
