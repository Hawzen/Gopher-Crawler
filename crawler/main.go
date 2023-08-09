package main

import (
	"log"
	"net/http"
	url_operations "net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const MAX_DEPTH = 2
const MAX_PAGES_BUFFER = 100
const MAX_URLS_PER_PAGE_PER_DOMAIN = 10
const SPIDER_COUNT = 10

// This is a var but please don't change it :)
var SPIDER_NAMES = [...]string{
	"Black Widow",
	"Brown Recluse",
	"Hobo Spider",
	"Tarantula",
	"Jumping Spider",
	"Crab Spider",
	"Wolf Spider",
	"Orb Weaver",
	"Camel Spider",
	"Daddy Longlegs",
	"Garden Spider",
	"Funnel Web Spider",
	"Sac Spider",
	"Cellar Spider",
	"Fishing Spider",
	"Trapdoor Spider",
	"Golden Silk Orb-Weaver",
	"Redback Spider",
	"Mouse Spider",
	"Banana Spider",
	"Brazilian Wandering Spider",
	"Goliath Birdeater",
	"Sydney Funnel-Web Spider",
	"Mexican Redknee Tarantula",
	"Peacock Spider",
	"Zebra Spider",
	"White-tailed Spider",
	"Spitting Spider",
	"Bold Jumping Spider",
	"Brown Huntsman Spider",
	"Ghost Spider",
	"Long-jawed Orb Weaver",
	"Marbled Orb Weaver",
	"Net-casting Spider",
	"Water Spider",
	"Woodlouse Spider",
	"Trapdoor Spider",
	"Bird-dropping Spider",
	"Crab-like Spiny Orb Weaver",
	"Domino Spider",
	"False Black Widow",
	"Grass Spider",
	"Happy Face Spider",
	"Metallic Green Jumping Spider",
	"Pumpkin Spider",
	"Red Widow Spider",
	"Silver Argiope",
	"Tan Jumping Spider",
	"Walnut Orb Weaver",
	"Yellow Sac Spider",
}

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
}

type Spider struct {
	id   int
	name string
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run . <target_url>")
	}

	target_url := os.Args[1]
	log.Printf("Nest established; target %s", target_url)

	index := Index{
		inprogress_or_done_pages: make(map[URL]bool),
		pages_to_crawl:           make(chan Page, MAX_PAGES_BUFFER),
	}
	index.pages_to_crawl <- Page{url: target_url}

	// Create spiders
	for i := 0; i < SPIDER_COUNT; i++ {
		spider := Spider{
			id:   i,
			name: SPIDER_NAMES[i],
		}
		go spider.crawl(&index)
	}

	time.Sleep(15 * time.Second)

	log.Printf("Nest destroyed; pages conqured:")
	for key := range index.inprogress_or_done_pages {
		log.Println(key)
	}
	log.Printf("Totalling %d pages", len(index.inprogress_or_done_pages))

}

func (spider *Spider) crawl(index *Index) {
	log.Printf("%s:\tstarted crawling", spider.name)
	for {
		page_to_crawl, no_more_pages := spider.fetch_page(index)
		if no_more_pages {
			log.Printf("%s:\tcommitted seppuku", spider.name)
			return
		}

		related_pages := spider.crawl_page(&page_to_crawl)
		if related_pages == nil {
			continue
		}
		spider.add_pages(related_pages, index)
		log.Printf("%s:\tfinished crawling page %s, related pages count: %d", spider.name, page_to_crawl.url, len(related_pages))

		// Add current page to the DB
	}
}

func (spider *Spider) add_pages(related_pages map[URL]Page, index *Index) {
	for url, page := range related_pages {
		// If the page is already in the index then don't add it
		if index.inprogress_or_done_pages[url] {
			continue
		}
		index.pages_to_crawl <- page
	}
}

func (spider *Spider) fetch_page(index *Index) (Page, bool) {
	page_to_crawl := <-index.pages_to_crawl
	index.inprogress_or_done_pages[page_to_crawl.url] = true
	return page_to_crawl, false
}

func (spider *Spider) crawl_page(page *Page) map[URL]Page {
	if page.depth > MAX_DEPTH {
		return nil
	}

	resp, err := http.Get(page.url)
	if err != nil {
		log.Println(err)
		return nil
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

	return page.related_pages
}

func find_related_pages(doc *goquery.Document, current_page *Page) map[URL]Page {
	related_pages := make(map[URL]Page)
	url_domain_to_count := make(map[string]int)

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("href")

		// If the url is invalid then skip
		url, ok := validate_url(url, current_page.url)
		if !ok {
			return
		}

		// If the url is in extracted_pages already then skip
		if _, ok := related_pages[url]; ok {
			return
		}

		// If we're getting lotsa urls from the same domain then skip
		url, ok = validate_max_url_count_per_domain(url, url_domain_to_count)
		if !ok {
			return
		}

		// Figure out the domain of the url
		parsed_url, err := url_operations.Parse(url)
		if err != nil {
			log.Fatal(err)
		}
		parts := strings.Split(parsed_url.Hostname(), ".")
		domain := parts[len(parts)-2] + "." + parts[len(parts)-1]
		url_domain_to_count[domain] += 1
		if url_domain_to_count[domain] > MAX_URLS_PER_PAGE_PER_DOMAIN {
			return
		}

		new_page := Page{
			url:        url,
			is_crawled: false,
			time_found: time.Now(),
			depth:      current_page.depth + 1,
		}

		related_pages[url] = new_page
	})
	return related_pages
}

func validate_url(url string, current_url URL) (URL, bool) {
	// If link is relative then make it absolute
	if strings.HasPrefix(url, "/") {
		// Get base of current page
		base_url, err := url_operations.Parse(current_url)
		if err != nil {
			return "", false
		}

		url = base_url.Scheme + "://" + base_url.Host + url
	}

	// If link is not http or https then mark it as invalid
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", false
	}

	// Remove trailing slash
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}

	return url, true
}

func validate_max_url_count_per_domain(url URL, url_domain_to_count map[string]int) (URL, bool) {
	parsed_url, err := url_operations.Parse(url)
	if err != nil {
		return "", false
	}
	parts := strings.Split(parsed_url.Hostname(), ".")
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]
	url_domain_to_count[domain] += 1
	if url_domain_to_count[domain] > MAX_URLS_PER_PAGE_PER_DOMAIN {
		return "", false
	}
	return url, true
}
