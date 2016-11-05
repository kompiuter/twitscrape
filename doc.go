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
The archive is publicly available and can be searched through at:
https://twitter.com/search-advanced?lang=en

No authentication is required and the package can be run without any
prior configurations.

You can start scraping by creating a new instance of the Scrape struct and
calling its Tweets method, by providing a search term, a start date and an end date.

    scr := twitscrape.Scrape{}
    start, _ := time.Parse("01/02/2006", "11/10/2009")
    until, _ := time.Parse("01/02/2006", "11/11/2009")
    // fetch tweets between start and until dates, which contain hashtag #golang
    tweets, err := scr.Tweets("#golang", start, until)

Tweets returns a slice of Tweet, which is a struct with the following fields:

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


By default, the Tweets function will not log anything (such as a missing attribute when scraping).
To enable logging, pass a io.Writer into the Scrape struct initialization and logging
will be written to that writer:

    scr := ts.Scrape{Info: os.Stdout}


In order to better refine your search, you may use any Query Operator (as defined by Twitter)
in your search term. The query operators can be found here:
https://dev.twitter.com/rest/public/search#query-operators

    // Tweets by Dave Cheney
    tweets, err := scr.Tweets("#golang from:davecheney", startDate, untilDate)


Since a Twitter search is paginated by Twitter (to 20 Tweets), this library abuses the fact
more tweets are loaded via AJAX. More information can be found in a great blog post by Tom Dickinson:
http://tomkdickinson.co.uk/2015/01/scraping-tweets-directly-from-twitters-search-page-part-1/
*/
package twitscrape
