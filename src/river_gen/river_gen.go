package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	id "github.com/wtolson/go-taglib"
)

type artist struct {
	Name   string
	Albums map[string]*album
}

type album struct {
	Name string
	Year int
	// Tracks map[string]*track
	Tracks []*track
	Path   string
}

type track struct {
	Name   string
	Number int
	Length string
}

func main() {

	artists := make(map[string]*artist)

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("path to dir missing")
		os.Exit(1)
	}
	directory := flag.Args()[0]

	fileList := []string{}
	filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, "flac") {
			fileList = append(fileList, path)
		}
		return nil
	})

	for _, file := range fileList {
		meta, _ := id.Read(file)

		// Artists
		artistName := meta.Artist()
		var art *artist
		var ok bool
		art, ok = artists[artistName]
		if ok == false {
			art = &artist{
				Name: artistName,
			}

			artists[artistName] = art
		}

		// Albums
		albumName := meta.Album()
		var alb *album
		alb, ok = art.Albums[albumName]
		if art.Albums == nil {
			art.Albums = make(map[string]*album)
		}
		if ok == false {
			path, _ := filepath.Split(strings.TrimRight(file, "/"))
			path, _ = filepath.Split(strings.TrimLeft(path, directory))

			alb = &album{
				Name: albumName,
				Year: meta.Year(),
				Path: path,
			}
		}
		art.Albums[albumName] = alb

		// Tracks
		if alb.Tracks == nil {
			alb.Tracks = []*track{}
		}
		alb.Tracks = append(alb.Tracks, &track{
			Name:   meta.Title(),
			Number: meta.Track(),
			Length: durationToString(meta.Length()),
		})

		// trackName := meta.Title()
		// var tr *track
		// tr, ok = alb.Tracks[trackName]
		// if alb.Tracks == nil {
		// 	alb.Tracks = make(map[string]*track)
		// }
		// if ok == false {
		// 	tr = &track{
		// 		Name:   trackName,
		// 		Number: meta.Track(),
		// 	}
		// }
		// alb.Tracks[trackName] = tr

		meta.Close()
	}

	b, _ := json.Marshal(artists)
	// json.Marshal(artists)

	fmt.Println(string(b))

}

func durationToString(d time.Duration) string {

	seconds := int(d.Seconds())

	var hours = seconds / 3600
	if hours != 0 {
		seconds -= 3600 * hours
	}

	var minutes = seconds / 60

	if minutes != 0 {
		seconds -= 60 * minutes
	}

	if hours != 0 {
		return fmt.Sprintf("%0d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)

}
