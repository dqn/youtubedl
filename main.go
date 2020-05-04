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

const mb = 1 << 20

type PlayerResponse struct {
	StreamingData StreamingData `json:"streamingData"`
	Microformat   Microformat   `json:"microformat"`
}

type Format struct {
	URL      string `json:"url"`
	MimeType string `json:"mimeType"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Quality  string `json:"quality"`
	Bitrate  int    `json:"bitrate"`
}

type AdaptiveFormat struct {
	Format
	Fps            int `json:"fps"`
	AverageBitrate int `json:"averageBitrate"`
}

type StreamingData struct {
	Formats         []Format         `json:"formats"`
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

type Progress struct {
	numerator   int64
	denominator int64
}

func (p *Progress) Write(b []byte) (int, error) {
	n := len(b)
	p.numerator += int64(n)

	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rdownloading... %.1f", float64(p.numerator)/mb)
	if p.denominator > 0 {
		fmt.Printf("/%.1f", float64(p.denominator)/mb)
	}
	fmt.Print(" MB")

	return n, nil
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

func printVideoInfo(videoID string, r *PlayerMicroformatRenderer, f *Format) {
	videoURL := "https://www.youtube.com/watch?v=" + videoID
	fmt.Printf("Channel: %s (%s)\n", r.OwnerChannelName, r.OwnerProfileURL)
	fmt.Printf("Title: %s (%s)\n", r.Title.SimpleText, videoURL)
	fmt.Printf("Published: %s\n", r.PublishDate)
	fmt.Printf("Length: %ss\n", r.LengthSeconds)
	fmt.Printf("View Count: %s views\n", r.ViewCount)
	fmt.Printf("Mime Type: %s\n", f.MimeType)
	fmt.Printf("Size: %dx%d\n", f.Width, f.Height)
	fmt.Printf("Quality: %s\n", f.Quality)
	fmt.Printf("Bitrate: %d bps\n", f.Bitrate)
}

func findExtention(s string) string {
	return s[strings.Index(s, "/")+1 : strings.Index(s, ";")]
}

func download(u, dest string) error {
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	w := &Progress{denominator: resp.ContentLength}
	_, err = io.Copy(f, io.TeeReader(resp.Body, w))
	fmt.Println()
	if err != nil {
		return err
	}

	return nil
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

	var highestQuality Format
	for _, v := range p.StreamingData.Formats {
		if v.Bitrate > highestQuality.Bitrate {
			highestQuality = v
		}
	}

	pmr := p.Microformat.PlayerMicroformatRenderer
	printVideoInfo(videoID, &pmr, &highestQuality)
	fmt.Println()

	ext := findExtention(highestQuality.MimeType)
	dest := fmt.Sprintf("%s.%s", pmr.Title.SimpleText, ext)
	if err := download(highestQuality.URL, dest); err != nil {
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
