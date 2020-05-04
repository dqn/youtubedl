package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type PlayerResponse struct {
	StreamingData StreamingData `json:"streamingData"`
	Microformat   Microformat   `json:"microformat"`
}

type AdaptiveFormat struct {
	URL            string `json:"url"`
	MimeType       string `json:"mimeType"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	Quality        string `json:"quality"`
	Fps            int    `json:"fps"`
	AverageBitrate int    `json:"averageBitrate"`
}

type StreamingData struct {
	AdaptiveFormats []AdaptiveFormat `json:"adaptiveFormats"`
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
	fmt.Println("Usage: youtubedl <video id>")
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

func printVideoInfo(videoID string, r *PlayerMicroformatRenderer, f *AdaptiveFormat) {
	videoURL := "https://www.youtube.com/watch?v=" + videoID
	fmt.Printf("Channel: %s (%s)\n", r.OwnerChannelName, r.OwnerProfileURL)
	fmt.Printf("Title: %s (%s)\n", r.Title.SimpleText, videoURL)
	fmt.Printf("Published: %s\n", r.PublishDate)
	fmt.Printf("Length: %ss\n", r.LengthSeconds)
	fmt.Printf("View Count: %s views\n", r.ViewCount)
	fmt.Printf("Mime Type: %s\n", f.MimeType)
	fmt.Printf("FPS: %dfps\n", f.Fps)
	fmt.Printf("Size: %dx%d\n", f.Width, f.Height)
	fmt.Printf("Quality: %s\n", f.Quality)
	fmt.Printf("Average Bitrate: %d bps\n", f.AverageBitrate)
}

func findExtention(s string) string {
	return s[strings.Index(s, "/")+1 : strings.Index(s, ";")]
}

func run() error {
	if len(os.Args) != 2 {
		usage()
		os.Exit(2)
	}

	videoID := os.Args[1]
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

	var highestQuality AdaptiveFormat
	for _, v := range p.StreamingData.AdaptiveFormats {
		if v.AverageBitrate > highestQuality.AverageBitrate {
			highestQuality = v
		}
	}

	printVideoInfo(videoID, &p.Microformat.PlayerMicroformatRenderer, &highestQuality)
	fmt.Println()
	fmt.Println("downloading...")

	resp, err := http.Get(highestQuality.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ext := findExtention(highestQuality.MimeType)
	dst := fmt.Sprintf("%s.%s", p.Microformat.PlayerMicroformatRenderer.Title.SimpleText, ext)
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
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
