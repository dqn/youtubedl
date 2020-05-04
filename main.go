package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type PlayerResponse struct {
	StreamingData StreamingData `json:"streamingData"`
	Microformat   Microformat   `json:"microformat"`
}

type AdaptiveFormat struct {
	URL              string `json:"url"`
	MimeType         string `json:"mimeType"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	LastModified     string `json:"lastModified"`
	ContentLength    string `json:"contentLength"`
	Quality          string `json:"quality"`
	Fps              int    `json:"fps"`
	QualityLabel     string `json:"qualityLabel"`
	ApproxDurationMs string `json:"approxDurationMs"`
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

	videoURL := "https://www.youtube.com/watch?v=" + videoID
	pmr := p.Microformat.PlayerMicroformatRenderer
	fmt.Printf("Channel: %s (%s)\n", pmr.OwnerChannelName, pmr.OwnerProfileURL)
	fmt.Printf("Title: %s (%s)\n", pmr.Title.SimpleText, videoURL)
	fmt.Printf("Published: %s\n", pmr.PublishDate)
	fmt.Printf("Length: %ss\n", pmr.LengthSeconds)
	fmt.Printf("View Count: %s\n", pmr.ViewCount)

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
