package domain

// A spoof has the same shape as the Article, but has different semantic meaning.
// Because of the similarity, Article.ToSpoof is a method on Article to easily make spoofs.
type Spoof Article

type SpoofStub struct {
	Slug     string
	Title    string
	Subtitle string
}
