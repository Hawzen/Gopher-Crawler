package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	url_operations "net/url"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"

	"github.com/PuerkitoBio/goquery"
	"github.com/jedib0t/go-pretty/v6/table"
)

// Config
const MAX_DEPTH = 30
const MAX_URL_PER_PAGE = 5
const MAX_URLS_PER_PAGE_PER_DOMAIN = 5

const SPIDER_COUNT = 5
const MAX_PAGES_BUFFER = 10000
const CRAWL_TIME = 70 * time.Second

// Analyzer config
const PORT = "9898"
const ANALYZER_URL = "http://localhost:" + PORT

var ANALYZER_ENDPOINTS = map[string]string{
	"keywords": ANALYZER_URL + "/keywords",
	"summary":  ANALYZER_URL + "/summarize",
}

// This is a var but please don't change it :)
var SPIDER_NAMES = [...]string{
	"Black Widow        🖤",            // The female black widow spider is known for eating the male after mating.
	"Brown Recluse      👀",            // The brown recluse spider's venom can cause necrosis, or tissue death.
	"Hobo Spider        🎒",            // The hobo spider is also known as the aggressive house spider.
	"Tarantula          🕸️",           // Tarantulas can regenerate lost limbs.
	"Jumping Spider     🦗",            // Jumping spiders have excellent vision and can jump up to 50 times their body length.
	"Crab Spider    `   🦀",            // Crab spiders are named for their crab-like appearance and movement.
	"Wolf Spider        🐺",            // Wolf spiders are known for their hunting ability and maternal care.
	"Orb Weaver         🕸️",           // Orb weaver spiders spin large, circular webs.
	"Camel Spider       🐫",            // Camel spiders are not true spiders and are actually a type of solifugae.
	"Daddy Longlegs     🦵",            // Daddy longlegs are not actually spiders, but are arachnids.
	"Garden Spider      🌻",            // Garden spiders are also known as writing spiders due to the zigzag pattern in their webs.
	"Funnel Web Spider  🕳️",           // Funnel web spiders are known for their venomous bite.
	"Sac Spider         🛍️",           // Sac spiders are named for their habit of building a sac-like web for shelter.
	"Cellar Spider      🍷",            // Cellar spiders are also known as daddy longlegs spiders and are often found in dark, damp places.
	"Fishing Spider     🎣",            // Fishing spiders are able to walk on water and hunt aquatic prey.
	"Trapdoor Spider    🚪",            // Trapdoor spiders build burrows with a hinged door made of silk and soil.
	"Golden Silk Orb-Weaver 🌟",        // Golden silk orb-weavers are known for their golden silk, which is stronger than steel.
	"Redback Spider     🔴",            // Redback spiders are venomous and are found in Australia.
	"Mouse Spider       🐭",            // Mouse spiders are named for their burrowing behavior and not their prey.
	"Banana Spider      🍌",            // Banana spiders are also known as golden orb-weavers and are found in the Americas.
	"Brazilian Wandering Spider 🚶",    // Brazilian wandering spiders are venomous and are known for their wandering behavior.
	"Goliath Birdeater  🐦",            // Goliath birdeaters are the largest spiders in the world by mass.
	"Sydney Funnel-Web Spider 🇦🇺",     // Sydney funnel-web spiders are venomous and are found in Australia.
	"Mexican Redknee Tarantula 🇲🇽",    // Mexican redknee tarantulas are popular as pets and are known for their docile nature.
	"Peacock Spider     🦚",            // Peacock spiders are known for their colorful, iridescent markings and elaborate courtship dances.
	"Zebra Spider       🦓",            // Zebra spiders are known for their black and white stripes and their ability to jump long distances.
	"White-tailed Spider 🦌",           // White-tailed spiders are found in Australia and are known for their venomous bite.
	"Spitting Spider    💦",            // Spitting spiders are able to spit venom at their prey from a distance.
	"Bold Jumping Spider 🕺",           // Bold jumping spiders are known for their bright colors and their ability to jump long distances.
	"Brown Huntsman Spider 🌳",         // Brown huntsman spiders are found in Australia and are known for their large size and speed.
	"Ghost Spider       👻",            // Ghost spiders are named for their pale, translucent appearance.
	"Long-jawed Orb Weaver 🕸️",        // Long-jawed orb weavers are known for their long, thin jaws and their ability to catch small insects.
	"Marbled Orb Weaver 🕸️",           // Marbled orb weavers are known for their colorful markings and their ability to spin large webs.
	"Net-casting Spider 🕸️",           // Net-casting spiders are able to catch prey by throwing a web over them.
	"Water Spider       🌊",            // Water spiders are able to walk on water and hunt aquatic prey.
	"Woodlouse Spider   🐞",            // Woodlouse spiders are named for their habit of preying on woodlice.
	"Bird-dropping Spider 💩",          // Bird-dropping spiders are able to camouflage themselves as bird droppings to avoid predators.
	"Crab-like Spiny Orb Weaver 🦀",    // Crab-like spiny orb weavers are named for their crab-like appearance and spiny legs.
	"Domino Spider      🎲",            // Domino spiders are named for their black and white markings.
	"False Black Widow  🕸️",           // False black widows are often mistaken for black widows, but are not venomous.
	"Grass Spider       🌿",            // Grass spiders are named for their habit of building webs in grassy areas.
	"Happy Face Spider  😊",            // Happy face spiders are named for the smiley face pattern on their abdomen.
	"Metallic Green Jumping Spider 💚", // Metallic green jumping spiders are known for their bright green color and their ability to jump long distances.
	"Pumpkin Spider     🎃",            // Pumpkin spiders are named for their orange color and their habit of building webs in pumpkin patches.
	"Red Widow Spider   👩🔴",           // Red widow spiders are venomous and are found in the southern United States.
	"Silver Argiope     🕸️",           // Silver argiopes are known for their large, silver webs and their habit of eating their webs each night.
	"Tan Jumping Spider 🦶",            // Tan jumping spiders are known for their tan color and their ability to jump long distances.
	"Walnut Orb Weaver  🥜",            // Walnut orb weavers are named for their habit of building webs in the shape of a walnut.
	"Yellow Sac Spider  💛",            // Yellow sac spiders are venomous and are often found in homes.
	"Emerald Spider    💚",             // Emerald spiders are named for their bright green color.
	"Golden Web Spinner 🏵️",           // Golden web spinners are known for their golden silk and their ability to spin large webs.
	"Ruby Hunter       💎",             // Ruby hunters are named for their bright red color and their hunting ability.
	"Sapphire Stalker  🌀",             // Sapphire stalkers are named for their bright blue color and their hunting ability.
	"Topaz Trapper     🌟",             // Topaz trappers are named for their bright yellow color and their ability to trap prey.
	"Amethyst Ambusher 💜",             // Amethyst ambushers are named for their bright purple color and their hunting ability.
	"Jade Jumper       🍏",             // Jade jumpers are named for their bright green color and their ability to jump long distances.
	"Opal Orb Weaver   🌐",             // Opal orb weavers are named for their iridescent markings and their ability to spin large webs.
	"Quartz Creeper    🕸️",            // Quartz creepers are named for their habit of blending in with rocks and other minerals.
	"Turquoise Tarantula 🏝️",          // Turquoise tarantulas are named for their bright blue color and are found in the Caribbean.
	"Bronze Biter      🥉",             // Bronze biters are named for their metallic bronze color and their hunting ability.
	"Silver Spinner    🥈",             // Silver spinners are named for their metallic silver color and their ability to spin webs.
	"Golden Gobbler    🥇",             // Golden gobblers are named for their golden color and their habit of eating their webs each night.
	"Platinum Pouncer  🏆",             // Platinum pouncers are named for their metallic platinum color and their hunting ability.
	"Titanium Trapper  🛡️",            // Titanium trappers are named for their metallic titanium color and their ability to trap prey.
	"Nickel Nester     🌰",             // Nickel nesters are named for their metallic nickel color and their habit of building nests.
	"Copper Catcher    🥉",             // Copper catchers are named for their metallic copper color and their ability to catch prey.
	"Zinc Zapper       ⚡",             // Zinc zappers are named for their metallic zinc color and their ability to move quickly.
	"Aluminum Attacker 🥊",             // Aluminum attackers are named for their metallic aluminum color and their hunting ability.
	"Iron Invader      🗡️",            // Iron invaders are named for their metallic iron color and their aggressive behavior.
	"Lead Leaper       🏃",             // Lead leapers are named for their metallic lead color and their ability to jump long distances.
	"Tin Tracker       🕵️",            // Tin trackers are named for their metallic tin color and their ability to track prey.
	"Steel Stalker     🗡️",            // Steel stalkers are named for their metallic steel color and their hunting ability.
	"Magnesium Mover   🚲",             // Magnesium movers are named for their metallic magnesium color and their ability to move quickly.
	"Potassium Pursuer 🍌",             // Potassium pursuers are named for their bright yellow color and their hunting ability.
	"Sodium Sprinter   🧂",             // Sodium sprinters are named for their metallic sodium color and their ability to move quickly.
	"Calcium Crawler   🥛",             // Calcium crawlers are named for their metallic calcium color and their habit of crawling.
	"Chlorine Chaser   🏊",             // Chlorine chasers are named for their metallic chlorine color and their ability to move quickly.
	"Argon Ambusher    🌬️",            // Argon ambushers are named for their inert nature and their hunting ability.
	"Helium Hopper     🎈",             // Helium hoppers are named for their lightness and their ability to jump long distances.
	"Hydrogen Hunter   💧",             // Hydrogen hunters are named for their abundance in the universe and their hunting ability.
	"Oxygen Orb Weaver 🌬️",            // Oxygen orb weavers are named for their importance in respiration and their ability to spin large webs.
	"Carbon Catcher    🖤",             // Carbon catchers are named for their importance in life and their ability to catch prey.
	"Neon Nester       🌈",             // Neon nesters are named for their bright neon color and their habit of building nests.
	"Silicon Stalker   🏝️",            // Silicon stalkers are named for their abundance in the earth's crust and their hunting ability.
	"Phosphorus Pursuer 🔥",            // Phosphorus pursuers are named for their importance in life and their hunting ability.
	"Sulfur Spinner    💨",             // Sulfur spinners are named for their distinctive smell and their ability to spin webs.
	"Potassium Pouncer 🍌",             // Potassium pouncers are named for their bright yellow color and their hunting ability.
	"Vanadium Vaulter  🏎️",            // Vanadium vaulters are named for their metallic vanadium color and their ability to jump long distances.
	"Chromium Creeper  🌈",             // Chromium creepers are named for their metallic chromium color and their habit of crawling.
	"Manganese Mover   🕸️",            // Manganese movers are named for their metallic manganese color and their ability to move quickly.
	"Iron Invader      🔨",             // Iron invaders are named for their metallic iron color and their aggressive behavior.
	"Cobalt Catcher    🔵",             // Cobalt catchers are named for their metallic cobalt color and their ability to catch prey.
	"Nickel Nester     🥈",             // Nickel nesters are named for their metallic nickel color and their habit of building nests.
	"Copper Creeper    🥉",             // Copper creepers are named for their metallic copper color and their habit of crawling.
	"Zinc Zapper       ⚡",             // Zinc zappers are named for their metallic zinc color and their ability to move quickly.
	"Gallium Grabber   🌡️",            // Gallium grabbers are named for their metallic gallium color and their ability to grab prey.
	"Germanium Gnasher 💎",             // Germanium gnashers are named for their metallic germanium color and their hunting ability.
	"Arsenic Ambusher  ☠️",            // Arsenic ambushers are named for their toxicity and their hunting ability.
	"Selenium Sprinter 🌞",             // Selenium sprinters are named for their metallic selenium color and their ability to move quickly.
	"Bromine Biter     🔥",             // Bromine biters are named for their toxicity and their hunting ability.
	"Krypton Kicker    🌬️",            // Krypton kickers are named for their abundance in the universe and their ability to move quickly.
	"Rubidium Runner   🔋",             // Rubidium runners are named for their metallic rubidium color and their ability to run long distances.
	"Strontium Stalker 💀",             // Strontium stalkers are named for their metallic strontium color and their hunting ability.
	"Yttrium Yanker    🌈",             // Yttrium yankers are named for their metallic yttrium color and their ability to grab prey.
	"Zirconium Zipper  💍",             // Zirconium zippers are named for their metallic zirconium color and their ability to spin webs.
	"Niobium Nibbler   🍔",             // Niobium nibblers are named for their metallic niobium color and their habit of nibbling.
	"Molybdenum Mover  🏔️",            // Molybdenum movers are named for their metallic molybdenum color and their ability to move quickly.
	"Technetium Trapper 🕸️",           // Technetium trappers are named for their rarity and their ability to trap prey.
	"Ruthenium Runner  🏃",             // Ruthenium runners are named for their metallic ruthenium color
	"Bagheera Kiplingi 🌱",             // Herbivorous spider
	"Portia Spider    🧠",              // Intelligent hunting spider
	"Swedish Spider   🇸🇪",             // Carl Alexander Clerck's spider
	"Linnaeus Spider  🕷️",             // Linnaeus' spider
	"Simon Spider     🍷",              // Eugène Simon's spider
	"Platnick Spider  📚",              // Norman Platnick's spider
	"Levi Spider      📖",              // Herbert Walter Levi's spider
	"Strand Spider    🌊",              // Embrik Strand's spider
	"Thorell Spider   🧵",              // Tamerlan Thorell's spider
	"Violin Spider    🎻",              // Brown Recluse alternate name
	"Dangerous Spider ☠️",             // Brown Recluse's potential danger to humans
}

type URL = string

type Page struct {
	UID           string    `json:"uid,omitempty"`
	URL           URL       `json:"url,omitempty"`
	Domain        Domain    `json:"domain,omitempty"`
	Title         string    `json:"title,omitempty"`
	Related_pages []Page    `json:"related_pages,omitempty"`
	Is_crawled    bool      `json:"is_crawled,omitempty"`
	Depth         uint      `json:"depth,omitempty"`
	Time_crawled  time.Time `json:"time_crawled,omitempty"`
	Time_found    time.Time `json:"time_found,omitempty"`
	DType         []string  `json:"dgraph.type,omitempty"`
	Summary       string    `json:"summary,omitempty"`
	Keywords      []*string `json:"keywords,omitempty"`
	related_pages map[URL]Page
}

type Domain struct {
	UID  string `json:"uid,omitempty"`
	Name string `json:"name,omitempty"`
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

	dg := Db_setup()

	target_url := os.Args[1]
	log.Infof("Nest established; target %s", target_url)

	index := Index{
		inprogress_or_done_pages: make(map[URL]*Page),
		pages_to_crawl:           make(chan Page, MAX_PAGES_BUFFER),
	}
	index.pages_to_crawl <- Page{
		URL: target_url,
	}

	// Create spiders
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
		go spider.crawl(&index, dg)
	}

	time.Sleep(CRAWL_TIME)

	log.Infof("Nest destroyed; pages conqured:")
	display_crawled_pages(&index)

	log.Infof("Totalling %d pages", len(index.inprogress_or_done_pages))

}

// ## Spider functions

func (spider *Spider) crawl(index *Index, dg *dgo.Dgraph) {
	spider.logger.Infof("started crawling")
	for {
		page_to_crawl := spider.fetch_page(index)

		related_pages := spider.crawl_page(page_to_crawl)
		if related_pages != nil {
			spider.add_related_pages(page_to_crawl, index)
		}
		spider.logger.Info("finished crawling page", "URL", page_to_crawl.URL, "related pages count", len(related_pages), "index", len(index.pages_to_crawl))

		spider.add_page_to_db(page_to_crawl, dg)
	}
}

func (spider *Spider) add_related_pages(page *Page, index *Index) {
	for _, related_page := range page.related_pages {
		if page.URL == related_page.URL {
			continue
		}

		select {
		case index.pages_to_crawl <- related_page:
		// If the channel is full then skip
		default:
			return
		}
	}
}

func (spider *Spider) fetch_page(index *Index) *Page {
	for {
		page_to_crawl := <-index.pages_to_crawl
		// If the page is already inprogress or done then skip
		if _, ok := index.inprogress_or_done_pages[page_to_crawl.URL]; ok {
			continue
		}
		index.inprogress_or_done_pages[page_to_crawl.URL] = &page_to_crawl
		return &page_to_crawl
	}
}

func (spider *Spider) crawl_page(page *Page) map[URL]Page {
	if page.Depth > MAX_DEPTH {
		return nil
	}

	resp, err := http.Get(page.URL)
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

	// Get page summary and keywords
	html, err := doc.Html()
	if err != nil {
		spider.logger.Warn(err)
		return nil
	}
	page.Summary = get_summary(page, html)
	page.Keywords = get_keywords(page, html)

	// Get page domain
	parsed_url, err := url_operations.Parse(page.URL)
	if err != nil {
		spider.logger.Warn(err)
		return nil
	}

	parts := strings.Split(parsed_url.Hostname(), ".")
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]
	page.Domain = Domain{
		Name: domain,
	}

	// Extract all links from the page
	page.related_pages = find_related_pages(doc, page)

	// Mark the page as crawled
	page.Title = doc.Find("title").Text()
	page.Time_crawled = time.Now()
	page.Is_crawled = true

	return page.related_pages
}

func (spider *Spider) add_page_to_db(page *Page, dg *dgo.Dgraph) {
	// Add current page to the DB # how you know
	Related_page := []Page{}
	for _, p := range page.related_pages {
		Related_page = append(Related_page, p)
	}
	page.Related_pages = Related_page

	// Create a new transaction
	txn := dg.NewTxn()

	// Create a new request
	req := &api.Request{CommitNow: true}

	// Create a new query
	req.Query = `query {
		page(func: eq(url, "` + page.URL + `")) {
			v as uid
		}

		domain(func: eq(name, "` + page.Domain.Name + `")) {
			d as uid
		}
	}
	`

	// Create a new mutation
	page.UID = "uid(v)"
	page.Domain.UID = "uid(d)"

	// Marshal the new Page node into a JSON byte array
	newPageBytes, err := json.Marshal(page)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new mutation
	mu := &api.Mutation{
		SetJson: newPageBytes,
	}

	// Add the mutation to the request
	req.Mutations = []*api.Mutation{mu}

	// Execute the query
	// Update email only if matching uid found.
	if _, err := dg.NewTxn().Do(context.Background(), req); err != nil {
		log.Fatal(err, "query", req.Query)
	}

	spider.logger.Info("adding page to db", "URL", page.URL)

	// Commit the transaction
	err = txn.Commit(context.Background())
}

func get_summary(page *Page, html string) string {
	resp, err := http.Post("http://localhost:9898/summarize", "application/json", bytes.NewBuffer([]byte(html)))
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}

func get_keywords(page *Page, html string) []*string {
	resp, err := http.Post(ANALYZER_ENDPOINTS["keywords"], "application/json", bytes.NewBuffer([]byte(html)))
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var keywords []*string
	err = json.Unmarshal(body, &keywords)
	if err != nil {
		return []*string{}
	}

	return keywords
}

// ## Page functions

func find_related_pages(doc *goquery.Document, current_page *Page) map[URL]Page {
	related_pages := make(map[URL]Page)
	url_domain_to_count := make(map[string]int)

	count_added := 0
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if count_added > MAX_URL_PER_PAGE {
			return
		}

		url, _ := s.Attr("href")

		// If the url is invalid then skip
		url, ok := validate_url(url, current_page.URL)
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
			log.Debug("skipping url", "URL", url, "reason", "too many urls from the same domain", "domain", domain)
			return
		}

		new_page := Page{
			URL:        url,
			Is_crawled: false,
			Time_found: time.Now(),
			Depth:      current_page.Depth + 1,
		}

		related_pages[url] = new_page
		count_added += 1
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

// ## Misc functions

func display_crawled_pages(index *Index) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"URL", "Title", "Depth", "Number of related pages"})
	for url, page := range index.inprogress_or_done_pages {
		if page.Is_crawled == false {
			continue
		}
		t.AppendRow(table.Row{url, page.Title, page.Depth, len(page.related_pages)})
	}
	t.SetTitle("Crawled pages")
	t.Render()
}
