# GopherCrawler

## What
This is a webcrawler that attempts to create a symantic graph of a given target url

## How
- Use Go to crawl the target url
  - Visit [this](./docs/crawler.md) for more info
- Use Dgraph to store the data
  - Visit [this](./docs/dgraph.md) for more info
- Use Flask to analyze the data
  - Visit [this](./docs/analyzer.md) for more info

# Run

## Need
- docker
- go


## Command
```
docker run --name dgraph -d -p "8181:8080" -p "9080:9080" -v dgraph-data:/dgraph dgraph/standalone:latest
docker run --name ratel  -d -p "8000:8000"  dgraph/ratel:latest
cd analyzer
pip install -r requirements.txt
python server.py & # Or open in a new terminal, without the & at the end
cd ../crawler
go run main.go <target_url>
```

- Browse `http://localhost:8000/`
- Execute query to see graph
```graphql
{ 
  Page(func: eq(is_crawled, "true")) {
		title
    summary
    keywords
    related_pages {
			url
    }
  }
  
  domain(func: has(name)) {
		name
  }
}
```