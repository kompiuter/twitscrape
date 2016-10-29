/*
MIT License
Copyright (c) 2016 Kyriacos Kyriacou
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

/*
Package twitscrape is a library for scraping tweets from the twitter archive.
The archive is pubicly available and can be searched through at:
https://twitter.com/search-advanced?lang=en

No authentication is required and the package can be run without any
prior configurations.

Query operators may be used on the search term using the standard query
operators as defined by Twitter:
https://dev.twitter.com/rest/public/search#query-operators
*/
package twitscrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Tweet represents each individual tweet retrieved from the archive
type Tweet struct {
	// Link to tweet in the form https://www.twitter.com/user/status
	Permalink string
	// Screen name (twitter handle) of tweet author
	Name string
	// Timestamp of tweet in UTC
	Timestamp time.Time
	// Contents of tweet
	Contents string
	// Tweet ID
	ID string
}

var (
	errNoTweets = errors.New("tweets: no tweets found")
	// minID is the ID of the minimum tweet.
	// The minimum tweet is the first tweet returned from the first scrape. It should
	// only be set once.
	minID string
)

// Tweets searches the Twitter archive and returns all tweets found
// between the start and until date. If you have provided a large period of time,
// this may take some time to compute. Feedback will be written to Out if you set
// Verbose to true (package level)
//
// Any query operator may be used in the search string to refine your search, as defined by Twitter:
// https://dev.twitter.com/rest/public/search#query-operators
func Tweets(search string, start, until time.Time) ([]Tweet, error) {
	// maxID tracks the current maximum tweet ID that we currently have.
	// It should be updated each time a scrape is performed as the last tweet received
	// by that scrape
	var maxID string
	// f encapsulates logic for formatting the URL
	f := func(maxID string) string {
		const searchf = "https://twitter.com/i/search/timeline?f=tweets&vertical=default&q=%s since:%s until:%s&src=typd"
		const df = "2006-01-02"
		u := fmt.Sprintf(searchf, url.QueryEscape(search), start.Format(df), until.Format(df))
		u = strings.Replace(u, " ", "%20", -1)
		// On first call, we don't know any tweet ID's so the query 'max position' will not be added.
		// On subsequents calls maxID should never be empty.
		if maxID != "" {
			u += fmt.Sprintf("&max_position=TWEET-%s-%s", maxID, minID)
		}
		return u
	}

	// t will hold tweets as they are coming in from scrapes
	var t []Tweet
loop:
	for {
		// Twitter public search only returns top 20 tweets, we need to loop
		// until we catch them all :).
		// See doc.go for more information.
		u := f(maxID)
		tw, err := tweets(u)
		switch err {
		case errNoTweets: // no more tweets, can stop looping
			break loop
		case nil: // all good!
		default:
			return nil, err
		}
		// minID is the absolute minimum tweet ID for this request, so it should only bet set once
		if minID == "" {
			minID = tw[0].ID
		}
		t = append(t, tw...)
		// maxID is the maximum tweet ID that we have so far
		maxID = tw[len(tw)-1].ID
		if maxID == minID { // need to check to avoid infinite loop (i.e. only 1 tweet returned)
			break loop
		}
	}
	return t, nil
}

// tweets returns all tweets scraped from the given url
func tweets(url string) (t []Tweet, err error) {
	html, err := getHTML(url)
	if err != nil {
		return nil, fmt.Errorf("tweets: %v", err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("tweets: %v", err)
	}
	sel := doc.Find(".tweet.js-stream-tweet.js-actionable-tweet.js-profile-popup-actionable.original-tweet.js-original-tweet")
	if sel.Nodes == nil {
		return nil, errNoTweets
	}

	var tweets []Tweet
	// ATTENTION: this selector iterator must always execute FIRST (before the two following)
	// It is responsible for initially creating the tweet structs.
	// Scrapes permalink and from it derive screen name & tweet id
	sel.Each(func(i int, s *goquery.Selection) {
		const statusf = "https://www.twitter.com%s"
		p, ok := s.Attr("data-permalink-path")
		if !ok {
			logf("tweet %d: could not get permalink\n", i)
			tweets = append(tweets, Tweet{}) // create empty tweet so that timestamp scraping doesn't fail
			return
		}
		// p is in form '/user/status/tweetid'
		sl := strings.Split(p, "/")
		if len(sl) < 4 {
			logf("tweet %d: permalink %s was not in correct format\n", i, p)
			tweets = append(tweets, Tweet{}) // create empty tweet so that timestamp scraping doesn't fail
			return
		}
		tweets = append(tweets, Tweet{Permalink: fmt.Sprintf(statusf, p), Name: sl[1], ID: sl[3]})
	})

	// Scrapes timestamp
	doc.Find(".tweet-timestamp.js-permalink.js-nav.js-tooltip").Each(func(i int, s *goquery.Selection) {
		t, ok := s.Attr("title")
		if !ok {
			logf("tweet %d: could not get timestamp\n", i)
			return
		}
		if i > len(tweets) { // should never occur
			logf("timestamp: found %d timestamps, only %d tweets exist\n", i, len(tweets))
			return
		}
		tme, err := time.Parse("3:04 PM - 2 Jan 2006", t)
		if err != nil {
			tme = time.Time{}
			logf("tweet %d: timestamp: could not parse time %s\n", i, t)
		}
		tweets[i].Timestamp = tme
	})

	// Scrapes contents of tweet
	doc.Find(".js-tweet-text-container").Each(func(i int, s *goquery.Selection) {
		t := strings.Trim(s.Text(), " \n")
		if t == "" {
			logf("tweet %d: contents were empty\n", i)
			return
		}
		if i > len(tweets) { // should never occur
			logf("text: found %d contents, only %d tweets exist\n", i, len(tweets))
			return
		}
		tweets[i].Contents = t
	})

	logf("%d tweets processed\n", len(tweets))
	return tweets, nil
}

// getHTML returns the HTML body of the JSON response that is returned by calling the Twitter
// advanced search URL
func getHTML(url string) (string, error) {
	logf("fetching %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	r := struct {
		HTML string `json:"items_html"`
	}{}
	err = dec.Decode(&r)
	if err != nil {
		return "", fmt.Errorf("could not decode: %v", err)
	}

	return r.HTML, nil
}

// Out will have info written to if verbose is set to true.
// Default is os.Stdout.
var Out io.Writer = os.Stdout

// Verbose will enable writing of info to io.Writer if set to true.
// Default is false.
var Verbose bool

// logf is a wrapper for fmt.Fprintf which writes to Out if Verbose is set to true
func logf(format string, a ...interface{}) {
	if Verbose {
		fmt.Fprintf(Out, format, a...)
	}
}
