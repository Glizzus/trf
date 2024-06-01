package scraping

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/glizzus/trf/internal/domain"
)

// GoqueryScraper is a scraper that is implemented using the goquery library.
type GoqueryScraper struct{}

func (s *GoqueryScraper) docFromURL(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create http request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to execute http request: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response body into document: %w", err)
	}

	return doc, nil
}

// LatestFactChecks returns the slugs of the latest fact checks.
// Snopes lays out its fact checks in the following order:
//
// # Newest
//
// # Second newest
//
// ...
//
// # Oldest
//
// Therefore, we return the slice in in this order:
//
//	[newest, second newest, ..., oldest].
func (s *GoqueryScraper) LatestFactChecks(ctx context.Context) (slugs []string, err error) {
	const baseURL = "https://www.snopes.com/fact-check/"

	doc, err := s.docFromURL(ctx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get document for latest fact checks: %w", err)
	}

	elements := doc.Find(".article_wrapper > .outer_article_link_wrapper")
	slugs = make([]string, elements.Length())
	elements.Each(func(i int, s *goquery.Selection) {
		articleURL, ok := s.Attr("href")
		if !ok {
			slog.Warn("No href found for latest fact check", "element", s)
			return
		}
		slug := strings.TrimSuffix(strings.TrimPrefix(articleURL, baseURL), "/")
		slugs[i] = slug
	})

	return slugs, nil
}

func extractDate(container *goquery.Selection) (date time.Time, err error) {
	dateString := container.Find(".publish_date").Text()
	if dateString == "" {
		return date, fmt.Errorf("could not find date")
	}
	dateString = strings.TrimPrefix(dateString, "Published ")

	date, err = time.Parse("January 2, 2006", dateString)
	if err != nil {
		return date, fmt.Errorf("could not parse date %s: %w", dateString, err)
	}

	return date, nil
}

func extractRating(container *goquery.Selection) (rating domain.Rating, err error) {
	var ratingStr string
	container.Find(".rating_title_wrap").Contents().EachWithBreak(func(i int, s *goquery.Selection) bool {
		if goquery.NodeName(s) == "#text" {
			ratingStr = strings.TrimSpace(s.Text())
			return false
		}
		return true
	})
	if ratingStr == "" {
		return rating, fmt.Errorf("could not find rating")
	}

	rating, err = domain.ParseRating(ratingStr)
	if err != nil {
		return rating, fmt.Errorf("could not parse rating: %w", err)
	}

	return rating, nil
}

func extractClaim(doc *goquery.Document) (claim domain.Claim, err error) {
	factCheckContainer := doc.Find("#fact_check_rating_container")

	question := factCheckContainer.Find(".claim_cont").Text()
	question = strings.TrimSpace(question)
	if question == "" {
		return claim, fmt.Errorf("could not find question")
	}
	claim.Question = strings.TrimSpace(question)

	rating, err := extractRating(factCheckContainer)
	if err != nil {
		return claim, fmt.Errorf("could not extract rating: %w", err)
	}
	claim.Rating = rating

	context := factCheckContainer.Find(".fact_check_info_description").Text()
	if context == "" {
		claim.Context = nil
	} else {
		claim.Context = &context
	}
	claim.Question = doc.Find(".claim_cont").Text()

	if claim.Question == "" {
		return claim, fmt.Errorf("no claim found")
	}

	return claim, nil
}

func (s *GoqueryScraper) ScrapeArticle(ctx context.Context, slug string) (article domain.Article, err error) {
	doc, err := s.docFromURL(ctx, "https://www.snopes.com/fact-check/"+slug)
	if err != nil {
		return article, fmt.Errorf("unable to get document for article %s: %w", slug, err)
	}

	titleContainer := doc.Find("section.title-container")

	article.Title = titleContainer.Find("h1").Text()
	if article.Title == "" {
		return article, fmt.Errorf("no title found for article %s", doc.Url)
	}

	article.Subtitle = titleContainer.Find("h2").Text()
	if article.Subtitle == "" {
		return article, fmt.Errorf("no subtitle found for article %s", doc.Url)
	}

	date, err := extractDate(titleContainer)
	if err != nil {
		return article, fmt.Errorf("could not extract date: %w", err)
	}
	article.Date = date

	claim, err := extractClaim(doc)
	if err != nil {
		return article, fmt.Errorf("could not extract claim: %w", err)
	}
	article.Claim = claim

	article.Content = scrapeContent(doc.Find("#article-content"))
	article.Slug = slug

	return article, nil
}

func scrapeContent(s *goquery.Selection) []string {
	var content []string
	s.Children().Each(func(i int, s *goquery.Selection) {
		if s.Is("section") {
			return
		}
		if s.Is("script") {
			return
		}
		if s.Is("input") {
			return
		}

		if s.Is("p") {
			text := extractText(s)
			if text != "" {
				content = append(content, text)
			}
		} else {
			// If it isn't any of the above, it is probably a div or something similar.
			// We will recurse into it. Note that we don't keep the recursive structure,
			// we just put it into the same flat slice.
			content = append(content, scrapeContent(s)...)
		}
	})
	return content
}

func extractText(n *goquery.Selection) string {
	var texts []string

	n.Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#text" {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				texts = append(texts, text)
			}
		} else {
			// If it isn't a text node, it is probably a span or something similar.
			// We will recurse into it.
			texts = append(texts, extractText(s))
		}
	})

	return strings.Join(texts, " ")
}
