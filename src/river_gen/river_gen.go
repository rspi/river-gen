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

type response struct {
	Sorted   map[string]*artist
	Unsorted *folder
}

type artist struct {
	Name   string
	Albums map[string]*album
}

type album struct {
	Name   string
	Year   int
	Tracks []*track
	Path   string
}

type track struct {
	Name   string
	Number int
	Length string
	Path   string
}

type folder struct {
	Name    string
	Files   []*file
	Folders map[string]*folder
}

func (f *folder) addFile(segments []string) {
	if len(segments) > 1 {
		subFolder, ok := f.Folders[segments[0]]
		if !ok {
			subFolder = newFolder(segments[0])
		}
		f.Folders[segments[0]] = subFolder
		subFolder.addFile(segments[1:])
	}

	if len(segments) == 1 {
		f.Files = append(f.Files, &file{segments[0]})
	}
}

func newFolder(name string) *folder {
	return &folder{
		Name:    strings.TrimRight(name, "/"),
		Files:   []*file{},
		Folders: make(map[string]*folder),
	}
}

type file struct {
	Name string
}

func main() {

	sortedPath := flag.String("sorted", "", "Directory to parse by music file tags")
	unsortedPath := flag.String("unsorted", "", "Directory to parse by folder structure")
	flag.Parse()

	if *sortedPath == "" && *unsortedPath == "" {
		fmt.Println("You need to provide at least one type of path. See -help")
		os.Exit(1)
	}

	resp := &response{
		Sorted:   createByTags(*sortedPath),
		Unsorted: createByPath(*unsortedPath),
	}

	output, _ := json.Marshal(resp)
	fmt.Println(string(output))
}

func createByPath(dir string) *folder {
	root := newFolder("Root")

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		path = strings.TrimLeft(strings.TrimRight(path, "/"), dir)
		segments := strings.Split(path, string(filepath.Separator))
		if strings.HasSuffix(path, "flac") {
			root.addFile(segments)
		}
		return nil
	})

	return root
}

func createByTags(dir string) map[string]*artist {
	artists := make(map[string]*artist)
	fileList := []string{}
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
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
			path, _ = filepath.Split(strings.TrimLeft(path, dir))

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
			Path:   strings.TrimLeft(file, dir),
		})
		meta.Close()
	}

	return artists
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
