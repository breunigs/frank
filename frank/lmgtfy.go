package frank

import (
	irc "github.com/fluffle/goirc/client"
	"log"
	"net/url"
	"regexp"
)

const googUrl = "http://googl.com/search?btnI=1&q="

// regex that matches lmgtfy requests
var lmgtfyMatcher = regexp.MustCompile(`^(?:[\d\pL._-]+: )?lmgtfy: (.+)`)

func Lmgtfy(conn *irc.Conn, line *irc.Line) {
	tgt := line.Args[0]
	msg := line.Args[1]

	if tgt[0:1] != "#" {
		// only answer to this in channels
		return
	}

	if !lmgtfyMatcher.MatchString(msg) {
		return
	}

	match := lmgtfyMatcher.FindStringSubmatch(msg)

	if len(match) < 2 {
		log.Printf("WTF: lmgtfy regex match didn’t have enough parts")
		return
	}

	u := googUrl + url.QueryEscape(match[1])
	t, lastUrl, err := TitleGet(u)

	if err != nil {
		conn.Privmsg(tgt, "[LMGTFY] "+lastUrl)
	} else {
		conn.Privmsg(tgt, "[LMGTFY] "+t+" @ "+lastUrl)
	}
}
