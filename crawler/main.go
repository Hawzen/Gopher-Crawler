package main

import (
	"net/http"
	url_operations "net/url"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	"github.com/PuerkitoBio/goquery"
	"github.com/jedib0t/go-pretty/v6/table"
)

const MAX_DEPTH = 3
const MAX_PAGES_BUFFER = 100000
const MAX_URLS_PER_PAGE_PER_DOMAIN = 10
const SPIDER_COUNT = 10
const CRAWL_TIME = 20 * time.Second

// This is a var but please don't change it :)
var SPIDER_NAMES = [...]string{
	"Black Widow        🖤",
	"Brown Recluse      👀",
	"Hobo Spider        🎒",
	"Tarantula          🕸️",
	"Jumping Spider     🦗",
	"Crab Spider        🦀",
	"Wolf Spider        🐺",
	"Orb Weaver         🕸️",
	"Camel Spider       🐫",
	"Daddy Longlegs     🦵",
	"Garden Spider      🌻",
	"Funnel Web Spider  🕳️",
	"Sac Spider         🛍️",
	"Cellar Spider      🍷",
	"Fishing Spider     🎣",
	"Trapdoor Spider    🚪",
	"Golden Silk Orb-Weaver 🌟",
	"Redback Spider     🔴",
	"Mouse Spider       🐭",
	"Banana Spider      🍌",
	"Brazilian Wandering Spider 🚶",
	"Goliath Birdeater  🐦",
	"Sydney Funnel-Web Spider 🇦🇺",
	"Mexican Redknee Tarantula 🇲🇽",
	"Peacock Spider     🦚",
	"Zebra Spider       🦓",
	"White-tailed Spider 🦌",
	"Spitting Spider    💦",
	"Bold Jumping Spider 🕺",
	"Brown Huntsman Spider 🌳",
	"Ghost Spider       👻",
	"Long-jawed Orb Weaver 🕸️",
	"Marbled Orb Weaver 🕸️",
	"Net-casting Spider 🕸️",
	"Water Spider       🌊",
	"Woodlouse Spider   🐞",
	"Bird-dropping Spider 💩",
	"Crab-like Spiny Orb Weaver 🦀",
	"Domino Spider      🎲",
	"False Black Widow  🕸️",
	"Grass Spider       🌿",
	"Happy Face Spider  😊",
	"Metallic Green Jumping Spider 💚",
	"Pumpkin Spider     🎃",
	"Red Widow Spider   👩🔴",
	"Silver Argiope     🕸️",
	"Tan Jumping Spider 🦶",
	"Walnut Orb Weaver  🥜",
	"Yellow Sac Spider  💛",
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
	inprogress_or_done_pages map[URL]*Page
	pages_to_crawl           chan Page
}

type Spider struct {
	id     int
	name   string
	logger *log.Logger
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run . <target_url>")
	}

	target_url := os.Args[1]
	log.Infof("Nest established; target %s", target_url)

	index := Index{
		inprogress_or_done_pages: make(map[URL]*Page),
		pages_to_crawl:           make(chan Page, MAX_PAGES_BUFFER),
	}
	index.pages_to_crawl <- Page{url: target_url}

	// Create spiders
	// TODO: use per-spider logger instead of global logger
	for i := 0; i < SPIDER_COUNT; i++ {
		spider := Spider{
			id:   i,
			name: SPIDER_NAMES[i],
			logger: log.NewWithOptions(os.Stderr, log.Options{
				ReportTimestamp: true,
				TimeFormat:      time.Kitchen,
				Prefix:          SPIDER_NAMES[i],
			}),
		}
		go spider.crawl(&index)
	}

	time.Sleep(CRAWL_TIME)

	log.Infof("Nest destroyed; pages conqured:")
	display_crawled_pages(&index)

	log.Infof("Totalling %d pages", len(index.inprogress_or_done_pages))

}

func (spider *Spider) crawl(index *Index) {
	spider.logger.Infof("started crawling")
	for {
		page_to_crawl := spider.fetch_page(index)

		related_pages := spider.crawl_page(page_to_crawl)
		if related_pages != nil {
			spider.add_pages(related_pages, index)
		}
		spider.logger.Info("finished crawling page", "URL", page_to_crawl.url, "related pages count", len(related_pages), "index", len(index.pages_to_crawl))

		// Add current page to the DB
	}
}

func (spider *Spider) add_pages(related_pages map[URL]Page, index *Index) {
	for _, page := range related_pages {
		index.pages_to_crawl <- page
	}
}

func (spider *Spider) fetch_page(index *Index) *Page {
	for {
		page_to_crawl := <-index.pages_to_crawl
		// If the page is already inprogress or done then skip
		if _, ok := index.inprogress_or_done_pages[page_to_crawl.url]; ok {
			continue
		}
		index.inprogress_or_done_pages[page_to_crawl.url] = &page_to_crawl
		return &page_to_crawl
	}
}

func (spider *Spider) crawl_page(page *Page) map[URL]Page {
	if page.depth > MAX_DEPTH {
		return nil
	}

	resp, err := http.Get(page.url)
	if err != nil {
		spider.logger.Warn(err)
		return nil
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		spider.logger.Warn(err)
		return nil
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

	// Convert to lowercase
	url = strings.ToLower(url)

	return url, true
}

func validate_max_url_count_per_domain(url URL, url_domain_to_count map[string]int) (URL, bool) {
	parsed_url, err := url_operations.Parse(url)
	if err != nil {
		return "", false
	}
	parts := strings.Split(parsed_url.Hostname(), ".")
	if len(parts) < 2 {
		return "", false
	}
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]
	url_domain_to_count[domain] += 1
	if url_domain_to_count[domain] > MAX_URLS_PER_PAGE_PER_DOMAIN {
		return "", false
	}
	return url, true
}

func display_crawled_pages(index *Index) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"URL", "Title", "Depth", "Number of related pages"})
	for url, page := range index.inprogress_or_done_pages {
		if page.is_crawled == false {
			continue
		}
		t.AppendRow(table.Row{url, page.title, page.depth, len(page.related_pages)})
	}
	t.SetTitle("Crawled pages")
	t.Render()
}
