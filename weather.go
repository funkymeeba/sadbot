// Copyright 2014 James McGuire. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

var directions = []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}

const (
	weatherUpdateQuery   = `INSERT INTO weather_location (nick, location) VALUES (?, ?) ON DUPLICATE KEY UPDATE nick=?, location=?;`
	weatherLocationQuery = `SELECT location FROM weather_location where nick=?;`
	openweathermapURL    = `http://api.openweathermap.org/data/2.5/weather`
)

type owmData struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp     float64 `json:"temp"`
		Pressure float64 `json:"pressure"`
		Humidity float64 `json:"humidity"`
		TempMin  float64 `json:"temp_min"`
		TempMax  float64 `json:"temp_max"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All float64 `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int     `json:"type"`
		ID      int     `json:"id"`
		Message float64 `json:"message"`
		Country string  `json:"country"`
		Sunrise int     `json:"sunrise"`
		Sunset  int     `json:"sunset"`
	} `json:"sys"`
	ID   int    `json:"id"`
	Name string `json:"name"`
	Cod  int    `json:"cod"`
}

func (o *owmData) String() string {
	conditions := []string{}
	for _, weather := range o.Weather {
		conditions = append(conditions, strings.Title(weather.Description))
	}
	weatherConds := strings.Join(conditions, ", ")
	tempC := o.Main.Temp - 273.15
	tempF := tempC*1.8 + 32.0
	windDir := getDirection(o.Wind.Deg)
	windSpeedMph := o.Wind.Speed * 2.23694
	windSpeedKph := o.Wind.Speed * 3.6
	return fmt.Sprintf("%s. %.1f °F / %.1f °C. Humidity %.0f%%. Wind from the %s at %.1f m/h / %.1f km/h. (%s)",
		weatherConds, tempF, tempC, o.Main.Humidity, windDir, windSpeedMph, windSpeedKph, o.Name)
}

func getDirection(deg float64) string {
	return directions[int(deg/22.5+.5)%16]
}

func fetchWeather(location string) (*owmData, error) {
	log.Printf("Querying openweathermap for %s", location)
	owm, err := url.Parse(openweathermapURL)
	if err != nil {
		return nil, err
	}
	v := owm.Query()
	v.Set("q", location)
	v.Set("APPID", config.OpenWeatherMapAPIKey)
	owm.RawQuery = v.Encode()
	resp, err := http.Get(owm.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var owmdata owmData
	err = json.NewDecoder(resp.Body).Decode(&owmdata)
	if err != nil {
		return nil, err
	}
	log.Printf("%+v\n", owmdata)
	return &owmdata, nil
}

func findLocation(nick string) (string, error) {
	rows, err := db.Query(weatherLocationQuery, nick)
	if err != nil {
		return "", err
	}
	var location string
	for rows.Next() {
		if err := rows.Scan(&location); err != nil {
			return "", err
		}
	}
	return location, nil
}

func updateLocation(nick, location string) error {
	_, err := db.Exec(weatherUpdateQuery, nick, location, nick, location)
	if err != nil {
		return err
	}
	return nil
}

func showWeather(conn *irc.Conn, line *irc.Line) {
	if !strings.HasPrefix(line.Text(), "!w") {
		return
	}
	location := strings.TrimSpace(strings.TrimPrefix(line.Text(), "!w"))

	target_nick := line.Nick
	targeted := false

	switch {
	case strings.HasPrefix(location, "@"):
		target_nick = strings.TrimSpace(strings.TrimPrefix(location, "@"))
		location = ""
		targeted = true
	case strings.HasPrefix(location, "set "):
		location = strings.TrimSpace(strings.TrimPrefix(location, "set "))
		log.Printf("Updating location for %s to %s", line.Nick, location)
		err := updateLocation(line.Nick, location)
		if err != nil {
			log.Println("Error updating location:", err)
		}
		result := fmt.Sprintf("%s: Your location has been updated to %s.", line.Nick, location)
		conn.Privmsg(line.Target(), result)
		return
	case strings.HasPrefix(location, "clear"):
		log.Printf("Clearing stored location for %s", line.Nick)
		err := updateLocation(line.Nick, "")
		if err != nil {
			log.Println("Error updating location:", err)
		}
		result := fmt.Sprintf("%s: Your location has been cleared in the database.", line.Nick)
		conn.Privmsg(line.Target(), result)
		return
	case strings.HasPrefix(location, "help"):
		result := fmt.Sprintf("%s: Check the weather! set will set your location (!w set San Francisco, CA),"+
			" clear will remove your stored location, @ will show the weather for another nick (!w @sadbox),"+
			" and help will show this message.", line.Nick)
		conn.Privmsg(line.Target(), result)
		return
	}

	if location == "" || targeted {
		var err error
		location, err = findLocation(target_nick)
		if err != nil {
			log.Println("Error fetching location from DB:", err)
			return
		}
		if location == "" {
			result := ""
			if targeted {
				result = fmt.Sprintf("%s: %s hasn't ever set a location.", line.Nick, target_nick)
			} else {
				result = fmt.Sprintf("%s: You need to specify a location at least once. (!w set San Francisco, CA)", line.Nick)
			}
			conn.Privmsg(line.Target(), result)
			return
		}
	}

	weatherdata, err := fetchWeather(location)
	if err != nil {
		result := fmt.Sprintf("%s: I can't seem to find anything for %s", line.Nick, location)
		conn.Privmsg(line.Target(), result)
		log.Println("Error fetching weather data...", err)
		return
	}
	result := fmt.Sprintf("%s: %s", line.Nick, weatherdata.String())
	conn.Privmsg(line.Target(), result)
}
