package twitscrape

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestScrapeTweetsLength(t *testing.T) {
	// Period and search term provided return the first tweets with the hashtag #golang.
	// There are 18 tweets and that number should never change.
	scr := Scrape{}
	search := "#golang"
	df := "01/02/2006"
	start, _ := time.Parse(df, "11/10/2009")
	until, _ := time.Parse(df, "11/11/2009")
	tweets, err := scr.Tweets(search, start, until)
	if err != nil {
		t.Error(err)
	}
	want := 18
	if len(tweets) != want {
		t.Errorf("len(Tweets(%s, %s, %s)) = %d, want:%d", search, start.Format(df), until.Format(df), len(tweets), want)
	}
}

func TestScrapeTweetContent(t *testing.T) {
	// Retrieve first tweet ever that contains hashtag #golang.
	// This is a tweet by duncanmak (https://www.twitter.com/duncanmak/status/5602929333).
	// This should never change.
	scr := Scrape{}
	search := "#golang"
	df := "01/02/2006"
	start, _ := time.Parse(df, "11/10/2009")
	until, _ := time.Parse(df, "11/11/2009")
	tweets, err := scr.Tweets(search, start, until)
	if err != nil {
		t.Error(err)
	}
	time, _ := time.Parse("3:04 PM - 2 Jan 2006", "3:26 PM - 10 Nov 2009")
	want := Tweet{
		ID:        "5602929333",
		Name:      "duncanmak",
		Permalink: "https://www.twitter.com/duncanmak/status/5602929333",
		Contents:  "Watching Rob Pike's talk on Google's new #golang language. A lot of his points remind me of ML systems, I wonder what's new?",
		Timestamp: time,
	}
	firstTweet := tweets[len(tweets)-1]
	if firstTweet != want {
		t.Errorf("got: %#v,\nwant: %#v", firstTweet, want)
	}
}

func TestScrapeTweetInfo(t *testing.T) {
	var b bytes.Buffer
	scr := Scrape{&b}
	search := "#golang"
	df := "01/02/2006"
	start, _ := time.Parse(df, "11/10/2009")
	until, _ := time.Parse(df, "11/11/2009")
	_, err := scr.Tweets(search, start, until)
	if err != nil {
		t.Error(err)
	}
	scanner := bufio.NewScanner(&b)
	scanner.Split(bufio.ScanLines)
	want := []string{`fetching https://twitter.com/i/search/timeline?f=tweets&vertical=default&q=%23golang%20since:2009-11-10%20until:2009-11-11&src=typd`,
		`18 tweets processed`,
		`fetching https://twitter.com/i/search/timeline?f=tweets&vertical=default&q=%23golang%20since:2009-11-10%20until:2009-11-11&src=typd&max_position=TWEET-5602929333-5603770675`}
	var errs []string
	for i, ln := range want {
		scanner.Scan()
		if ln != scanner.Text() {
			errs = append(errs, fmt.Sprintf("got: %s, want: %s", ln, want[i]))
		}
		if err := scanner.Err(); err != nil {
			t.Error(err)
		}
	}
	if len(errs) > 0 {
		s := strings.Join(errs, "\n")
		t.Error(errors.New(s))
	}
}
