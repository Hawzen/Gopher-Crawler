package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type URL = string

type Page struct {
	url           URL
	title         string
	related_pages map[URL]Page
	is_crawled    bool
	depth         uint
	time_crawled  time.Time
	time_found    time.Time
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing target URL")
		os.Exit(1)
	}

	target_url := os.Args[1]
	log.Printf("Begin crawling %s", target_url)

	// Create a new page
	page := Page{url: target_url}
	crawl_page(&page)

}

func crawl_page(page *Page) map[URL]Page {
	resp, err := http.Get(page.url)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	log.Printf("Crawling %s, Status: %s, Size: %d", page.url, resp.Status, resp.ContentLength)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Extract all links from the page
	page.related_pages = find_related_pages(doc, page)

	// Mark the page as crawled
	page.title = doc.Find("title").Text()
	page.time_crawled = time.Now()
	page.is_crawled = true

	log.Printf("Finished crawling %s, Title: %s, Related Pages: %d", page.url, page.title, len(page.related_pages))

	// Add current page to the DB
	// TODO: Add to DB

	return page.related_pages
}

func find_related_pages(doc *goquery.Document, current_page *Page) map[URL]Page {
	related_pages := make(map[URL]Page)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")

		// If the link is in extracted_pages already then skip
		if _, ok := related_pages[link]; ok {
			return
		}

		new_page := Page{
			url:        link,
			is_crawled: false,
			time_found: time.Now(),
			depth:      current_page.depth + 1,
		}

		related_pages[link] = new_page
	})
	return related_pages
}
