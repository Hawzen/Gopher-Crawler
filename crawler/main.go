package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const MAX_DEPTH = 3

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

type Index struct {
	inprogress_or_done_pages map[URL]bool
	pages_to_crawl           chan Page
	mu                       sync.Mutex
}

type Spider struct {
	id int
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run . <target_url>")
	}

	target_url := os.Args[1]
	log.Printf("Nest established; target %s", target_url)

	index := Index{
		inprogress_or_done_pages: make(map[URL]bool),
		pages_to_crawl:           make(chan Page),
	}

	spider := Spider{
		id: 1,
	}

	// Begin crawling
	go spider.crawl(&index)

	// Add the target url to the index
	index.pages_to_crawl <- Page{url: target_url}
}

func (spider *Spider) crawl(index *Index) {
	log.Printf("Spider %d started crawling", spider.id)
	page_to_crawl := spider.fetch_page(index)
	log.Printf("Spider %d finished fetching page %s", spider.id, page_to_crawl.url)
	related_pages := spider.crawl_page(&page_to_crawl)
	log.Printf("Spider %d finished crawling page %s, related pages count: %d", spider.id, page_to_crawl.url, len(related_pages))

}

func (spider *Spider) fetch_page(index *Index) Page {
	index.mu.Lock()
	defer index.mu.Unlock()
	page_to_crawl := <-index.pages_to_crawl
	index.inprogress_or_done_pages[page_to_crawl.url] = true
	return page_to_crawl
}

func (spider *Spider) crawl_page(page *Page) map[URL]Page {
	if page.depth > MAX_DEPTH {
		return nil
	}

	resp, err := http.Get(page.url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		// In the future I should add "FAILED TO CRAWL" pages
		log.Fatal(err)
	}

	// Extract all links from the page
	page.related_pages = find_related_pages(doc, page)

	// Mark the page as crawled
	page.title = doc.Find("title").Text()
	page.time_crawled = time.Now()
	page.is_crawled = true

	log.Printf("Finished crawling %s, Title: %s, Related Pages: %d", page.url, page.title, len(page.related_pages))

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
