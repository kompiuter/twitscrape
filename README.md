# twitscrape
A scraper written in Go that fetches tweets from the public Twitter search. This does **not** require authentication.

# Installation
Obrain the library using:

```bash
$ go get -u github.com/kompiuter/twitscrape
```

# About
This library uses the publicly available [Twitter Advanced Search](https://twitter.com/search-advanced?lang=en) to search for tweets, scrapes all tweets from the response and returns a slice of tweets.

Since this is a scraper, no authentication whatsoever is required. You can use it right out of the box!

# Usage
Create a `Scrape` struct and use its `Tweets` method. You can specify an `io.Writer` when creating the `Scrape` struct to enable logging of info messages to that writer. 
Leaving `Info` empty means the scraper will not log any info messages.

```go
scr := twitscrape.Scrape{Info: os.Stdout}
start, _ := time.Parse("01/02/2006", "11/10/2009")
until, _ := time.Parse("01/02/2006", "11/11/2009")
// fetch tweets between start and until dates, which contain hashtag #golang
tweets, err := scr.Tweets("#golang", start, until)
if err != nil {
  // Handle err
}
fmt.Print(tweets[0].Permalink) // https://www.twitter.com/repeatedly/status/5603770675
```

See a complete example which writes tweets to an output file [here](https://github.com/kompiuter/twitscrape/blob/master/example/main.go).


## Query Operators

You may use any [query operator](https://dev.twitter.com/rest/public/search#query-operators) as a search parameter to refine your results.

For example:

```go
tweets, err := scr.Tweets("#golang from:davecheney", start, until)
```

## Tests

There are tests available and can be run in the root directory using:

```bash
go test
```

They test that correct output is returned when querying old tweets, relying on the premise that Twitter will never ammend or delete them. If they do, ¯\\_(ツ)_/¯
### Roadmap

- [ ] Create doc.go
- [ ] Add a complete example using query operators
- [ ] Return tweets using a chan

**Feedback and pull requests are most welcome!**




