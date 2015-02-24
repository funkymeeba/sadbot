// Copyright 2014 James McGuire. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	irc "github.com/fluffle/goirc/client"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/html/charset"
)

var (
	config   Config
	urlRegex = regexp.MustCompile(`(?i)\b((?:https?://|www\d{0,3}[.]|[` +
		`a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+` +
		`\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s` + "`" + `!()\[` +
		`\]{};:'".,<>?«»“”‘’]))`)
	httpRegex      = regexp.MustCompile(`https?://.*`)
	findWhiteSpace = regexp.MustCompile(`\s+`)
	db             *sql.DB
	badWords       = make(map[string]*regexp.Regexp)
	titleStart     = []byte("<title>")
	titleEnd       = []byte("</title>")
)

const FREENODE = "irc.freenode.net"

type Config struct {
	Channel       string
	DBConn        string
	Nick          string
	Ident         string
	FullName      string
	FlickrAPIKey  string
	WolframAPIKey string
	IRCPass       string
	RebuildWords  bool
	Commands      []struct {
		Name string
		Text string
	}
	BadWords []struct {
		Word  string
		Query string
	}
}

// Try and grab the title for any URL's posted in the channel
func sendUrl(channel, unparsedURL string, conn *irc.Conn) {
	if !httpRegex.MatchString(unparsedURL) {
		unparsedURL = `http://` + unparsedURL
	}
	postedUrl, err := url.Parse(unparsedURL)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Fetching title for " + postedUrl.String() + " In channel " + channel)

	resp, err := http.Get(postedUrl.String())
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Println("http server return error.")
		return
	}
	respbody := []byte{}
	if resp.Header.Get("Content-Type") == "" {
		buf := make([]byte, 512)
		bufsize, err := resp.Body.Read(buf)
		if err != nil {
			log.Println("adding content type failed")
		}
		resp.Header.Set("Content-Type", http.DetectContentType(buf[:bufsize]))
		respbody = append(respbody, buf[:bufsize]...)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		log.Println("content-type is not text/html")
		return
	}

	utf8Body, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		log.Println("error converting page to utf8")
		return
	}
	restofbody, err := ioutil.ReadAll(io.LimitReader(utf8Body, 50000))
	if err != nil {
		log.Println("error reading posted link")
		return
	}
	respbody = append(respbody, restofbody...)
	titlestart := bytes.Index(respbody, titleStart)
	titleend := bytes.Index(respbody, titleEnd)
	if titlestart == -1 || titleend == -1 {
		return
	}
	title := respbody[titlestart+7 : titleend]
	title = bytes.TrimSpace(title)
	if len(title) == 0 || !utf8.Valid(title) {
		return
	}
	// Example:
	// Title: sadbox . org (at sadbox.org)
	formattedTitle := html.UnescapeString(string(title))
	formattedTitle = findWhiteSpace.ReplaceAllString(formattedTitle, " ")
	formattedTitle = fmt.Sprintf("Title: %s (at %s)", formattedTitle, postedUrl.Host)
	log.Println(formattedTitle)
	conn.Privmsg(channel, formattedTitle)
}

func logMessage(line *irc.Line, channel, message string) {
	_, err := db.Exec("insert into messages (Nick, Ident, Host, Src, Cmd, Channel,"+
		" Message, Time) values (?, ?, ?, ?, ?, ?, ?, ?)", line.Nick, line.Ident,
		line.Host, line.Src, line.Cmd, channel, message, line.Time)
	if err != nil {
		log.Println(err)
	}
	err = updateWords(line.Nick, message)
	if err != nil {
		log.Println(err)
	}
}

func checkForUrl(channel string, splitmessage []string, conn *irc.Conn) {
	urllist := []string{}
	numlinks := 0
NextWord:
	for _, word := range splitmessage {
		word = strings.TrimSpace(word)
		if urlRegex.MatchString(word) {
			for _, subUrl := range urllist {
				if subUrl == word {
					continue NextWord
				}
			}
			numlinks++
			if numlinks > 3 {
				break
			}
			urllist = append(urllist, word)
			go sendUrl(channel, word, conn)
		}
	}
}

// This function does all the dispatching for various commands
// as well as logging each message to the database
func handleMessage(conn *irc.Conn, line *irc.Line) {
	// This is so that the bot can properly respond to pm's
	var channel string
	if conn.Me().Nick == line.Args[0] {
		channel = line.Nick
	} else {
		channel = line.Args[0]
	}
	message := line.Args[1]
	splitmessage := strings.Split(message, " ")
	cmd := strings.TrimSpace(splitmessage[0])

	// sadbox-only commands!
	if line.Nick == "sadbox" {
		switch cmd {
		case "!dance":
			go dance(channel, conn)
		case "!audio":
			go conn.Privmsg(channel, "https://sadbox.org/static/stuff/audiophile.html")
		case "!cst":
			go conn.Privmsg(channel, "\u00039,13#CSTMASTERRACE")
		case "!chatter":
			go markov(channel, conn)
		}
	}

	// Special commands
	switch cmd {
	case "!haata":
		go haata(channel, conn)
	case "!search":
		go googSearch(channel, message, conn)
	case "!ask":
		go wolfram(channel, message, line.Nick, conn)
	case "!meebcast":
		var command string
		if len(splitmessage) >= 2 {
			command = strings.TrimSpace(splitmessage[1])
		}
		go meeba(channel, line.Nick, command, conn)
	}

	// Commands that are read in from the config file
	for _, command := range config.Commands {
		if cmd == command.Name {
			var response string
			if len(splitmessage) >= 2 {
				response = fmt.Sprintf("%s: %s", splitmessage[1], command.Text)
			} else {
				response = command.Text
			}
			go conn.Privmsg(channel, response)
		}
	}

	// This is what looks at each word and tries to figure out if it's a URL
	go checkForUrl(channel, splitmessage, conn)

	// Shove that shit in the database!
	go logMessage(line, channel, message)
}

func init() {
	log.Println("Starting sadbot")

	runtime.GOMAXPROCS(runtime.NumCPU())

	rand.Seed(time.Now().UTC().UnixNano())

	configfile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.NewDecoder(configfile).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	for _, word := range config.BadWords {
		badWords[word.Word] = regexp.MustCompile(word.Query)
	}

	log.Println("Loaded config file!")
	log.Printf("Joining channel %s", config.Channel)
	log.Printf("Nick: %s", config.Nick)
	log.Printf("Ident: %s", config.Ident)
	log.Printf("FullName: %s", config.FullName)

	log.Printf("Found %d commands", len(config.Commands))
	for index, command := range config.Commands {
		log.Printf("%d %s: %s", index+1, command.Name, command.Text)
	}
}

func main() {
	var err error
	db, err = sql.Open("mysql", config.DBConn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxIdleConns(100)
	db.SetMaxOpenConns(200)

	go makeMarkov()

	buildchan := make(chan os.Signal, 1)
	signal.Notify(buildchan, syscall.SIGUSR1)
	go func() {
		for _ = range buildchan {
			genTables()
		}
	}()

	ircConfig := irc.NewConfig(config.Nick, config.Ident, config.FullName)
	ircConfig.SSL = true
	ircConfig.Server = FREENODE
	ircConfig.Pass = config.Nick + ":" + config.IRCPass

	c := irc.Client(ircConfig)

	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			conn.Join(config.Channel)
			log.Println("Connected!")
		})
	quit := make(chan bool)

	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	c.HandleFunc(irc.PRIVMSG, handleMessage)
	c.HandleFunc(irc.ACTION, handleMessage)

	if err := c.Connect(); err != nil {
		log.Fatalln("Connection error: %s\n", err)
	}

	<-quit
}
