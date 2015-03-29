package frank

import (
	"code.google.com/p/go.net/html"
	frankconf "github.com/breunigs/frank/config"
	irc "github.com/fluffle/goirc/client"
	rss "github.com/jteeuwen/go-pkg-rss"
	"log"
	"strconv"
	"time"
)

// how often to check the feeds (in minutes)
const checkEvery = 3

// ignore all posts that are older than X minutes
const freshness = 90

// if there’s an error reading a feed, retry after X minutes
const retryAfter = 9

// how many items to show if there have been many updates in an interval
const maxItems = 2

// reference time: Mon Jan 2 15:04:05 -0700 MST 2006
const timeFormat1 = time.RFC1123Z
const timeFormat2 = "2006-01-02T15:04:05Z"
const timeFormat3 = "2006-01-02T15:04:05-07:00"

var conn *irc.Conn

var ignoreBefore = time.Now()

var rssHttpClient = HttpClientWithTimeout()

func Rss(connection *irc.Conn) {
	conn = connection
	// this feels wrong, the missing alignment making it hard to read.
	// Does anybody have a suggestion how to make this nice in go?
	//~ go pollFeed("#i3-test", "i3", timeFormat2, "http://code.stapelberg.de/git/i3/atom/?h=next")
	go pollFeed("#i3", "i3website", timeFormat2, "http://code.stapelberg.de/git/i3-website/atom/?h=master")
	go pollFeed("#i3", "i3faq", timeFormat1, "https://faq.i3wm.org/feeds/rss/")
}

func pollFeed(channel string, feedName string, timeFormat string, uri string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg:RSS: %v\n", r)
			time.Sleep(retryAfter * time.Minute)
			pollFeed(channel, feedName, timeFormat, uri)
		}
	}()

	if frankconf.Verbose {
		log.Printf("RSS %s: Setting up %s to post to %s \n", feedName, uri, channel)
	}

	// this will process all incoming new feed items and discard all that
	// are somehow erroneous or older than the threshold. It will directly
	// post any updates.
	itemHandler := func(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
		log.Printf("RSS: %d new item(s) in %s\n", len(newitems), feedName)

		postitems := []string{}

		for _, item := range newitems {
			pubdate, err := time.Parse(timeFormat, item.PubDate)
			// ignore items with unreadable date format
			if err != nil {
				log.Printf("RSS: WTF @ reading date for %s: %s (err: %v)\n", feedName, item.PubDate, err)
				continue
			}

			// ignore items that were posted before frank booted or are older
			// than “freshness” minutes
			if ignoreBefore.After(pubdate) {
				log.Printf("RSS %s: skipping posts made before booting (posted: %s, booted: %s)\n", feedName, pubdate, ignoreBefore)
				continue
			}
			if time.Since(pubdate) >= freshness*time.Minute {
				log.Printf("RSS %s: skipping non-fresh post (posted: %s, time_ago: %s)\n", feedName, pubdate, time.Since(pubdate))
				continue
			}

			url := ""
			if len(item.Links) > 0 {
				url = item.Links[0].Href
			}

			if url != "" && isRecentUrl(url) {
				if frankconf.Verbose {
					log.Printf("RSS %s: Skipping item because saved as recent URL (URL: %s)\n", feedName, url)
				}
				continue
			}

			if url != "" {
				addRecentUrl(url)
				url = " @ " + url
			}

			author := html.UnescapeString(item.Author.Name)
			title := html.UnescapeString(item.Title)

			if author == "" {
				postitems = appendIfMiss(postitems, "::"+feedName+":: "+title+url)
			} else {
				postitems = appendIfMiss(postitems, "::"+feedName+":: "+title+url+" (by "+author+")")
			}
		}

		cnt := len(postitems)

		// hide updates if they exceed the maxItems counter. If there’s only
		// one more item in the list than specified in maxItems, all of the
		// items will be printed – otherwise that item would be replaced by
		// a useless message that it has been hidden.
		if cnt > maxItems+1 {
			cntS := strconv.Itoa(cnt)
			maxS := strconv.Itoa(maxItems)
			msg := "::" + feedName + ":: had " + cntS + " updates, showing the latest " + maxS
			conn.Privmsg(channel, msg)
			postitems = postitems[cnt-maxItems : cnt]
		}

		// newer items appear first in feeds, so reverse them here to keep
		// the order in line with how IRC wprks
		for i := len(postitems) - 1; i >= 0; i -= 1 {
			conn.Privmsg(channel, postitems[i])
			log.Printf("RSS %s: posting %s\n", feedName, postitems[i])
		}
	}

	// create the feed listener/updater
	feed := rss.New(checkEvery, true, chanHandler, itemHandler)

	// check for updates infinite loop
	for {
		if frankconf.Verbose {
			t := feed.LastUpdate().Format(time.RFC3339)
			log.Printf("RSS %s: Updating now (previous update: %s, refresh ok: %s)\n", feedName, t, feed.CanUpdate())
		}

		if err := feed.FetchClient(uri, &rssHttpClient, nil); err != nil {
			log.Printf("RSS %s: Error for %s: %s\n", feedName, uri, err)
			time.Sleep(retryAfter * time.Minute)
			continue
		}

		<-time.After(time.Duration(feed.SecondsTillUpdate() * 1e9))
	}
}

// unused default handler
func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	log.Printf("RSS: %d new channel(s) in %s\n", len(newchannels), feed.Url)
}

// append string to slice only if it’s not already present.
func appendIfMiss(slice []string, s string) []string {
	for _, elm := range slice {
		if elm == s {
			if frankconf.Verbose {
				log.Printf("RSS: Not adding “%s” because it is already present\n", s)
			}
			return slice
		}
	}
	return append(slice, s)
}

// LIFO that stores the recent posted URLs.
// Used to avoid posting entries multiple times that have been
// erroneously detected as new by the RSS library.

var recent []string = make([]string, 50)
var recentIndex = 0

func addRecentUrl(url string) {
	recent[recentIndex] = url
	recentIndex += 1
	if len(recent) == recentIndex {
		recentIndex = 0
	}
}

func isRecentUrl(url string) bool {
	for _, a := range recent {
		if url == a {
			return true
		}
	}
	return false
}
