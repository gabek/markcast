package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func generateTracklistFromCue(cueFilePath string) string {
	file, err := os.Open(cueFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	newTrack := false

	var tracks []Track
	var track Track
	lineCounter := 0
	hasInitializedFirstTrack := false

	for scanner.Scan() {
		// Skip the first 5 lines of header stuff
		lineCounter++
		if lineCounter < 6 {
			continue
		}

		line := scanner.Text()

		if newTrack {
			track = Track{}
			newTrack = false
		}

		lineComponents := strings.SplitN(line, " ", 2)
		lineKey := strings.TrimLeft(lineComponents[0], "\t")
		lineValue := lineComponents[1]

		if lineKey == "TITLE" {
			hasInitializedFirstTrack = true
			track.Name = removeQuotes(lineValue)
		} else if lineKey == "PERFORMER" {
			track.Artist = removeQuotes(lineValue)
		} else if lineKey == "INDEX" {
			positionComponents := strings.SplitN(lineValue, " ", 2)
			track.Position = positionComponents[1]
		} else if string(line[0]) == "\t" && string(line[1]) != "\t" {
			newTrack = true
			// Don't add duplicates in a row
			if hasInitializedFirstTrack && (len(tracks) == 0 || !isTrackEqual(track, tracks[len(tracks)-1])) {
				tracks = append(tracks, track)
			}
		}
	}

	// Add the final track
	tracks = append(tracks, track)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	output := "\n"

	for _, track := range tracks {
		line := fmt.Sprintf("1. %s **%s** - %s\n", track.Position, track.Artist, track.Name)
		output += line
	}

	return output
}

type Track struct {
	Artist   string
	Name     string
	Position string
}

func isTrackEqual(t1 Track, t2 Track) bool {
	if t1.Name == t2.Name && t1.Artist == t2.Artist {
		return true
	}
	return false
}

func removeQuotes(quotedString string) string {
	unquotedString := strings.TrimLeft(quotedString, `"`)
	unquotedString = strings.TrimRight(unquotedString, `"`)
	return unquotedString
}
