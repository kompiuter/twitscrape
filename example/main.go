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
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	ts "github.com/kompiuter/twitscrape"
)

func main() {
	// Retrieve all tweets containing hashtag #golang from 10-Nov-2009 to 12-Nov-2009
	scr := ts.Scrape{Info: os.Stdout}
	start, _ := time.Parse("01/02/2006", "11/10/2009")
	until, _ := time.Parse("01/02/2006", "11/12/2009")
	tweets, err := scr.Tweets("#golang", start, until)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	f, err := os.Create("out1.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	sort.Sort(byTimestamp(tweets))
	printTweets(f, tweets)
	f.Close()

	// Retrieve all tweets containing hashtag #golang by @davecheney from 10-Nov-2010 to 10-Nov-2011
	start, _ = time.Parse("01/02/2006", "11/10/2010")
	until, _ = time.Parse("01/02/2006", "11/10/2011")
	tweets, err = scr.Tweets("#golang from:davecheney", start, until)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	f, err = os.Create("out2.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	sort.Sort(byTimestamp(tweets))
	printTweets(f, tweets)
	f.Close()
}

// byTimestamp satisfies the Sort.Interface interface
type byTimestamp []ts.Tweet

func (t byTimestamp) Len() int           { return len(t) }
func (t byTimestamp) Less(i, j int) bool { return t[i].Timestamp.Before(t[j].Timestamp) }
func (t byTimestamp) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

func printTweets(out io.Writer, tweets []ts.Tweet) {
	const format = "%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(out, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Timestamp", "Permalink", "Contents")
	fmt.Fprintf(tw, format, "----------------", "------------------------------", "------------------------------")
	for _, t := range tweets {
		fmt.Fprintf(tw, format, t.Timestamp.Format("2006-01-02 15:04"), strings.TrimPrefix(t.Permalink, "https://www.twitter.com/"), t.Contents)
	}
	tw.Flush()
}
