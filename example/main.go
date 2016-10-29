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
	"os"
	"strings"
	"text/tabwriter"
	"time"

	ts "github.com/kompiuter/twitscrape"
)

func main() {
	ts.Verbose = true
	start, _ := time.Parse("01/02/2006", "11/10/2009")
	until, _ := time.Parse("01/02/2006", "11/11/2009")
	tweets, err := ts.Tweets("#golang", start, until)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	printTweets(tweets)
}

func printTweets(tweets []ts.Tweet) {
	const format = "%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Timestamp", "Permalink", "Contents")
	fmt.Fprintf(tw, format,
		"----------------",
		"------------------------------",
		"------------------------------------------------------------------------------------------------------------------------------------------------")
	for _, t := range tweets {
		fmt.Fprintf(tw, format, t.Timestamp.Format("2006-01-02 15:04"), strings.TrimPrefix(t.Permalink, "https://www.twitter.com/"), t.Contents)
	}
	tw.Flush()
}
