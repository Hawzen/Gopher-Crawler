# GopherCrawler

## What
This is a webcrawler that attempts to create a symantic graph of a given target url

## How
- Use Go to crawl the target url
  - Visit [this](./docs/crawler.md) for more info
- Use Dgraph to store the data
  - Visit [this](./docs/dgraph.md) for more info
- (WIP) Use Flask to analyze the data

# Run

## Need
- docker
- go


## Command
```
docker run --name dgraph -d -p "8181:8080" -p "9080:9080" -v dgraph-data:/dgraph dgraph/standalone:latest
docker run --name ratel  -d -p "8000:8000"  dgraph/ratel:latest
go run main.go <target_url>
```

