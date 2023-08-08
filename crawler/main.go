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

	page.title = doc.Find("title").Text()

	// Extract all links from the page
	page.related_pages = make(map[URL]Page)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")

		// If the link is in extracted_pages already then skip
		if _, ok := page.related_pages[link]; ok {
			return
		}

		new_page := Page{
			url:        link,
			is_crawled: false,
			time_found: time.Now(),
			depth:      page.depth + 1,
		}

		page.related_pages[link] = new_page
	})

	// Mark the page as crawled
	page.is_crawled = true
	page.time_crawled = time.Now()

	log.Printf("Finished crawling %s, Title: %s, Related Pages: %d", page.url, page.title, len(page.related_pages))

	return page.related_pages
}
