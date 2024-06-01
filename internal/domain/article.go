package domain

import "time"

type Claim struct {
	Question string `json:"question"`
	Rating   Rating `json:"rating"`

	Context *string `json:"context,omitempty"`
}

type Article struct {
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Date     time.Time `json:"date"`
	Claim    Claim `json:"claim"`

	Content string `json:"content"`
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
