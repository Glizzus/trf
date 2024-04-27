package scrape

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
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *GoqueryScraper) LatestFactChecks(ctx context.Context) (slugs []string, err error) {
	const baseURL = "https://www.snopes.com/fact-check/"
	doc, err := s.docFromURL(ctx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get document for latest fact checks: %w", err)
	}
	const selector = ".article_wrapper > .outer_article_link_wrapper"
	elements := doc.Find(selector)
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

func (s *GoqueryScraper) ScrapeArticle(ctx context.Context, slug string) (*domain.Article, error) {
	doc, err := s.docFromURL(ctx, "https://www.snopes.com/fact-check/"+slug)
	if err != nil {
		return nil, err
	}
	var article domain.Article
	article.Slug = slug

	titleContainer := doc.Find("section.title-container")

	article.Title = titleContainer.Find("h1").Text()
	if article.Title == "" {
		return nil, fmt.Errorf("no title found for article %s", doc.Url)
	}

	article.Subtitle = titleContainer.Find("h2").Text()
	if article.Subtitle == "" {
		return nil, fmt.Errorf("no subtitle found for article %s", doc.Url)
	}

	dateString := strings.TrimSpace(titleContainer.Find(".publish_date").Text())
	if dateString == "" {
		return nil, fmt.Errorf("no date found for article %s", doc.Url)
	}
	article.Date, err = time.Parse("January 2, 2006", dateString)

	factCheckContainer := doc.Find("#fact_check_rating_container")
	question := factCheckContainer.Find(".claim_cont").Text()
	if question == "" {
		return nil, fmt.Errorf("could not find question")
	}
	question = strings.TrimSpace(question)
	article.Claim.Question = question

	var ratingStr string
	factCheckContainer.Find(".rating_title_wrap").Contents().EachWithBreak(func(i int, s *goquery.Selection) bool {
		if goquery.NodeName(s) == "#text" {
			ratingStr = strings.TrimSpace(s.Text())
			return false
		}
		return true
	})
	if ratingStr == "" {
		return nil, fmt.Errorf("could not find rating")
	}
	rating, err := domain.ParseRating(ratingStr)
	if err != nil {
		return nil, err
	}
	article.Claim.Rating = rating

	context := factCheckContainer.Find(".fact_check_info_description").Text()
	if context == "" {
		article.Claim.Context = nil
	} else {
		article.Claim.Context = &context
	}

	content := scrapeContent(doc.Find("#article-content"))
	article.Content = strings.Join(content, "\n")

	return &article, nil
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
			/* If it isn't any of the above, it is probably a div or something similar.
			We will recurse into it. Note that we don't keep the recursive structure,
			we just put it into the same flat slice. */
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
			/* If it isn't a text node, it is probably a span or something similar.
			We will recurse into it. */
			texts = append(texts, extractText(s))
		}
	})

	return strings.Join(texts, " ")
}
