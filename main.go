package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/dqn/godl"
)

type PlayerResponse struct {
	StreamingData StreamingData `json:"streamingData"`
	Microformat   Microformat   `json:"microformat"`
}

type Format struct {
	URL          string `json:"url"`
	MimeType     string `json:"mimeType"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Quality      string `json:"quality"`
	Bitrate      int    `json:"bitrate"`
	AudioQuality string `json:"audioQuality"`
}

type StreamingData struct {
	Formats         []Format `json:"formats"`
	AdaptiveFormats []Format `json:"adaptiveFormats"`
}

type Title struct {
	SimpleText string `json:"simpleText"`
}

type Description struct {
	SimpleText string `json:"simpleText"`
}

type PlayerMicroformatRenderer struct {
	Title            Title       `json:"title"`
	Description      Description `json:"description"`
	LengthSeconds    string      `json:"lengthSeconds"`
	OwnerProfileURL  string      `json:"ownerProfileUrl"`
	ViewCount        string      `json:"viewCount"`
	PublishDate      string      `json:"publishDate"`
	OwnerChannelName string      `json:"ownerChannelName"`
}

type Microformat struct {
	PlayerMicroformatRenderer PlayerMicroformatRenderer `json:"playerMicroformatRenderer"`
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  youtubedl <video id> [options]")
	fmt.Println("Options:")
	fmt.Println("  -m: Download as music")
	os.Exit(2)
}

func getVideoInfo(videoID string) (url.Values, error) {
	u := "https://www.youtube.com/get_video_info?video_id=" + videoID
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	v, err := url.ParseQuery(string(b))
	if err != nil {
		return nil, err
	}

	return v, nil
}

func printCommonInfo(videoID string, r *PlayerMicroformatRenderer, f *Format) {
	videoURL := "https://www.youtube.com/watch?v=" + videoID
	fmt.Printf("Channel: %s (%s)\n", r.OwnerChannelName, r.OwnerProfileURL)
	fmt.Printf("Title: %s (%s)\n", r.Title.SimpleText, videoURL)
	fmt.Printf("Published: %s\n", r.PublishDate)
	fmt.Printf("Length: %ss\n", r.LengthSeconds)
	fmt.Printf("View Count: %s views\n", r.ViewCount)
	fmt.Printf("Mime Type: %s\n", f.MimeType)
}

func printMusicInfo(videoID string, r *PlayerMicroformatRenderer, f *Format) {
	printCommonInfo(videoID, r, f)
}

func printVideoInfo(videoID string, r *PlayerMicroformatRenderer, f *Format) {
	printCommonInfo(videoID, r, f)
	fmt.Printf("Size: %dx%d\n", f.Width, f.Height)
	fmt.Printf("Quality: %s\n", f.Quality)
	fmt.Printf("Bitrate: %d bps\n", f.Bitrate)
}

func findExtention(s string) string {
	return s[strings.Index(s, "/")+1 : strings.Index(s, ";")]
}

func run() error {
	flag.Parse()

	var (
		videoID string
		music   bool
	)
	switch flag.NArg() {
	case 1:
		videoID = flag.Arg(0)
	case 2:
		videoID = flag.Arg(0)
		if flag.Arg(1) != "-m" {
			usage()
		}
		music = true
	default:
		usage()
	}

	v, err := getVideoInfo(videoID)
	if err != nil {
		return err
	}

	if status, ok := v["status"]; !ok || status[0] != "ok" {
		fmt.Println(v)
		return fmt.Errorf("failed to get video info")
	}

	var p PlayerResponse
	if err := json.Unmarshal([]byte(v["player_response"][0]), &p); err != nil {
		return err
	}

	pmr := p.Microformat.PlayerMicroformatRenderer
	var highestQuality Format
	if music {
		for _, v := range p.StreamingData.AdaptiveFormats {
			if v.AudioQuality == "" {
				continue
			}
			if v.Bitrate > highestQuality.Bitrate {
				highestQuality = v
			}
		}
		printMusicInfo(videoID, &pmr, &highestQuality)
	} else {
		for _, v := range p.StreamingData.Formats {
			if v.Bitrate > highestQuality.Bitrate {
				highestQuality = v
			}
		}
		printVideoInfo(videoID, &pmr, &highestQuality)
	}
	fmt.Println()

	ext := findExtention(highestQuality.MimeType)
	dest := fmt.Sprintf("%s.%s", strings.ReplaceAll(pmr.Title.SimpleText, "/", "-"), ext)
	if err := godl.Download(highestQuality.URL, dest, true); err != nil {
		return err
	}

	fmt.Println("completed!")

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
