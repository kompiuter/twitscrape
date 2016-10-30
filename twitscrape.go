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

// Scrape is responsible for scraping tweets from Twitter.
// If Info is set to a writer then log messages will be written to that writer,
// otherwise no log messages will be written
type Scrape struct {
	Info io.Writer
}

var errNoTweets = errors.New("tweets: no tweets found")

// Tweets searches the Twitter archive and returns all tweets found
// between the start and until date. If you have provided a large period of time,
// this may take some time to complete.
// Info messages will be written to Scrape.Info
//
// Any query operator may be used in the search string to refine your search, as defined by Twitter:
// https://dev.twitter.com/rest/public/search#query-operators
func (s Scrape) Tweets(search string, start, until time.Time) ([]Tweet, error) {
	// minID is the ID of the minimum tweet.
	// The minimum tweet is the first tweet returned from the first scrape. It should
	// only be set once.
	var minID string
	// maxID tracks the current maximum tweet ID that we currently have.
	// It should be updated each time a scrape is performed as the last tweet received
	// by that scrape
	var maxID string
	// f encapsulates logic for formatting the URL
	f := func(maxID string) (*url.URL, error) {
		const searchf = "https://twitter.com/i/search/timeline?f=tweets&vertical=default&q=%s since:%s until:%s&src=typd"
		const df = "2006-01-02"
		raw := fmt.Sprintf(searchf, url.QueryEscape(search), start.Format(df), until.Format(df))
		raw = strings.Replace(raw, " ", "%20", -1)
		// On first call, we don't know any tweet ID's so the query 'max position' will not be added.
		// On subsequents calls maxID should never be empty.
		if maxID != "" {
			raw += fmt.Sprintf("&max_position=TWEET-%s-%s", maxID, minID)
		}
		u, err := url.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %v", raw, err)
		}
		return u, nil
	}

	// t will hold tweets as they are coming in from scrapes
	var t []Tweet
loop:
	for {
		// Twitter public search only returns top 20 tweets, we need to loop
		// until we catch them all :).
		// See doc.go for more information.
		u, err := f(maxID)
		if err != nil {
			return nil, err
		}
		tw, err := s.tweets(u)
		if err != nil {
			if err == errNoTweets { // no more tweets, can stop looping
				break loop
			}
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
func (s Scrape) tweets(u *url.URL) (tweets []Tweet, err error) {
	html, err := s.getHTML(u)
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

	// ATTENTION: this selector iterator must always execute FIRST (before the two following)
	// It is responsible for initially creating the tweet structs.
	// Scrapes permalink and from it derive screen name & tweet id
	sel.Each(func(i int, sel *goquery.Selection) {
		const statusf = "https://www.twitter.com%s"
		p, ok := sel.Attr("data-permalink-path")
		if !ok {
			s.infof("tweet %d: could not get permalink\n", i)
			tweets = append(tweets, Tweet{}) // create empty tweet so that timestamp scraping doesn't fail
			return
		}
		// p is in form '/user/status/tweetid'
		sl := strings.Split(p, "/")
		if len(sl) < 4 {
			s.infof("tweet %d: permalink %s was not in correct format\n", i, p)
			tweets = append(tweets, Tweet{}) // create empty tweet so that timestamp scraping doesn't fail
			return
		}
		tweets = append(tweets, Tweet{Permalink: fmt.Sprintf(statusf, p), Name: sl[1], ID: sl[3]})
	})

	// Scrapes timestamp
	doc.Find(".tweet-timestamp.js-permalink.js-nav.js-tooltip").Each(func(i int, sel *goquery.Selection) {
		t, ok := sel.Attr("title")
		if !ok {
			s.infof("tweet %d: could not get timestamp\n", i)
			return
		}
		if i > len(tweets) { // should never occur
			s.infof("timestamp: found %d timestamps, only %d tweets exist\n", i, len(tweets))
			return
		}
		tme, err := time.Parse("3:04 PM - 2 Jan 2006", t)
		if err != nil {
			tme = time.Time{}
			s.infof("tweet %d: timestamp: could not parse time %s\n", i, t)
		}
		tweets[i].Timestamp = tme
	})

	// Scrapes contents of tweet
	doc.Find(".js-tweet-text-container").Each(func(i int, sel *goquery.Selection) {
		t := strings.Trim(sel.Text(), " \n")
		if t == "" {
			s.infof("tweet %d: contents were empty\n", i)
			return
		}
		if i > len(tweets) { // should never occur
			s.infof("text: found %d contents, only %d tweets exist\n", i, len(tweets))
			return
		}
		tweets[i].Contents = t
	})

	s.infof("%d tweets processed\n", len(tweets))
	return tweets, nil
}

// getHTML returns the HTML body of the JSON response that is returned by calling the Twitter
// advanced search URL
func (s Scrape) getHTML(u *url.URL) (string, error) {
	raw := u.String()
	s.infof("fetching %s\n", raw)
	resp, err := http.Get(raw)
	if err != nil {
		return "", fmt.Errorf("GET %s: %v", raw, err)
	}
	defer resp.Body.Close()

	var out struct {
		HTML string `json:"items_html"`
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return "", fmt.Errorf("could not decode: %v", err)
	}

	return out.HTML, nil
}

// infof is a wrapper for fmt.Fprintf which writes to Info
func (s Scrape) infof(format string, a ...interface{}) {
	if s.Info == nil {
		return
	}
	fmt.Fprintf(s.Info, format, a...)
}
