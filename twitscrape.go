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
	"errors"
	"fmt"
	"io"
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
}

// Tweets searches the Twitter archive and returns the top 20 (if they exist) tweets
// between the start and until date.
//
// Any query operator may be used in the search string to refine your search, as defined by Twitter:
// https://dev.twitter.com/rest/public/search#query-operators
//
// Set verbose to true to enable logging and will log to Out.
func Tweets(search string, start, until time.Time) ([]Tweet, error) {
	const searchf = "https://twitter.com/search?f=tweets&vertical=default&q=%s since:%s until:%s&src=typd"
	const df = "2006-01-02"
	URL := fmt.Sprintf(searchf, url.QueryEscape(search), start.Format(df), until.Format(df))
	URL = strings.Replace(URL, " ", "%20", -1)

	logf("fetching %s\n", URL)
	doc, err := goquery.NewDocument(URL)
	if err != nil {
		return nil, fmt.Errorf("tweets: %v", err)
	}
	sel := doc.Find(".tweet.js-stream-tweet.js-actionable-tweet.js-profile-popup-actionable.original-tweet.js-original-tweet") // class that holds part of a tweet
	if sel.Nodes == nil {
		return nil, errors.New("tweets: no tweets found")
	}

	var tweets []Tweet
	// ATTENTION: this selector iterator must always execute FIRST (before the two following)
	// It is responsible for initially creating the tweet structs.
	// Scrapes permalink and screen name
	sel.Each(func(i int, s *goquery.Selection) {
		const statusf = "https://www.twitter.com%s"
		p, ok := s.Attr("data-permalink-path")
		if !ok {
			logf("tweet %d: could not get permalink\n", i)
		}
		n, ok := s.Attr("data-screen-name")
		if !ok {
			logf("tweet %d: could not get screen name\n", i)
		}

		tw := Tweet{Permalink: fmt.Sprintf(statusf, p), Name: n}
		tweets = append(tweets, tw)
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
		tme, err := time.Parse("3:04 PM - 02 Jan 2006", t)
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

	return tweets, nil
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
