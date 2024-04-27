package domain

import "time"

type Claim struct {
	Question string
	Rating   Rating

	Context *string
}

type Article struct {
	Slug     string
	Title    string
	Subtitle string
	Date     time.Time
	Claim    Claim

	Content string
}

// ToSpoof converts an Article to a Spoof.
// The newContent parameter is the content of the spoofed article.
// Everything else is the same as the original article, except the rating is opposite.
func (a *Article) ToSpoof(newContent string) Spoof {
	return Spoof{
		Slug:     a.Slug,
		Title:    a.Title,
		Subtitle: a.Subtitle,
		Date:     a.Date,
		Claim: Claim{
			Question: a.Claim.Question,
			Rating:   a.Claim.Rating.Opposite(),
			Context:  a.Claim.Context,
		},
		Content: newContent,
	}
}
