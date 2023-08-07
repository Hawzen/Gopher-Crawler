Run DGraph
```
docker run --name dgraph -d -p "8181:8080" -p "9080:9080" -v dgraph-data:/dgraph dgraph/standalone:latest
docker run --name ratel  -d -p "8000:8000"  dgraph/ratel:latest
```