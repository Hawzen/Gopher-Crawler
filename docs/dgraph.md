Sample query to use in ratel to view the data
```graphql
{
  Page(func: eq(url, "<target_url>")) {   
    url
    title
    depth
    related_pages {
      title
      url
      depth
      related_pages {
        url
        title
        depth 
      }
    }
  }
}
```