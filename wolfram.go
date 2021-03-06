// Copyright 2014 James McGuire. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const wolframAPIUrl = `http://api.wolframalpha.com/v2/query`

// Wolfram|Alpha structs
type Wolfstruct struct {
	Success bool  `xml:"success,attr"`
	Pods    []Pod `xml:"pod"`
}

type Pod struct {
	Title   string `xml:"title,attr"`
	Text    string `xml:"subpod>plaintext"`
	Primary bool   `xml:"primary,attr"`
}

func wolfram(conn *irc.Conn, line *irc.Line) {
	if !strings.HasPrefix(line.Text(), "!ask") {
		return
	}
	query := strings.TrimSpace(strings.TrimPrefix(line.Text(), "!ask"))
	if query == "" {
		conn.Privmsg(line.Target(), "Example: !ask pi")
		return
	}
	log.Printf("Searching wolfram alpha for %s", query)
	wolf, err := url.Parse(wolframAPIUrl)
	if err != nil {
		log.Println(err)
		return
	}
	v := wolf.Query()
	v.Set("input", query)
	v.Set("appid", config.WolframAPIKey)
	wolf.RawQuery = v.Encode()
	resp, err := http.Get(wolf.String())
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	var wolfstruct Wolfstruct
	err = xml.NewDecoder(resp.Body).Decode(&wolfstruct)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("%+v\n", wolfstruct)
	if !wolfstruct.Success {
		conn.Privmsg(line.Target(), "I have no idea.")
		return
	}
	var interpretation string
	for _, pod := range wolfstruct.Pods {
		if pod.Title == "Input interpretation" {
			interpretation = pod.Title + ": " + pod.Text
		}
		if !pod.Primary {
			continue
		}
		log.Println(query)
		response := strings.Split(pod.Title+": "+pod.Text, "\n")
		var numlines int
		if len(response) > 3 {
			numlines = 3
		} else {
			numlines = len(response)
		}
		query = fmt.Sprintf("(In reponse to: <%s> %s)", line.Nick, query)
		if interpretation != "" {
			conn.Privmsg(line.Target(), interpretation)
		}
		if numlines == 1 {
			conn.Privmsg(line.Target(), response[0]+" "+query)
		} else {
			for _, message := range response[:numlines] {
				conn.Privmsg(line.Target(), message)
			}
			conn.Privmsg(line.Target(), query)
		}
		// Sometimes it returns multiple primary pods
		return
	}
	// If I couldn't find anything just give up...
	conn.Privmsg(line.Target(), "I have no idea.")
}
