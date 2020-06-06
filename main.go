package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dqn/godl"
	"github.com/dqn/ytvi"
)

func getBetween(str, a, b string) string {
	return str[strings.Index(str, a)+1 : strings.Index(str, b)]
}

func makeFileName(title, mime string) string {
	ext := getBetween(mime, "/", ";")
	return fmt.Sprintf("%s.%s", strings.ReplaceAll(title, "/", "-"), ext)
}

func downloadMusic(videoID string) error {
	p, err := ytvi.GetVideoInfo(videoID)
	if err != nil {
		return err
	}

	var f ytvi.AdaptiveFormat
	for _, v := range p.StreamingData.AdaptiveFormats {
		if v.AudioQuality == "" {
			continue
		}
		if v.Bitrate > f.Bitrate {
			f = v
		}
	}

	r := p.Microformat.PlayerMicroformatRenderer
	videoURL := "https://www.youtube.com/watch?v=" + videoID
	fmt.Printf("Channel: %s (%s)\n", r.OwnerChannelName, r.OwnerProfileURL)
	fmt.Printf("Title: %s (%s)\n", r.Title.SimpleText, videoURL)
	fmt.Printf("Published: %s\n", r.PublishDate)
	fmt.Printf("Length: %ss\n", r.LengthSeconds)
	fmt.Printf("View Count: %s views\n", r.ViewCount)
	fmt.Printf("Mime Type: %s\n", f.MimeType)
	fmt.Println()

	dest := makeFileName(r.Title.SimpleText, f.MimeType)

	return godl.Download(f.URL, dest, true)
}

func downloadVideo(videoID string) error {
	p, err := ytvi.GetVideoInfo(videoID)
	if err != nil {
		return err
	}

	var f ytvi.Format
	for _, v := range p.StreamingData.Formats {
		if v.Bitrate > f.Bitrate {
			f = v
		}
	}

	r := p.Microformat.PlayerMicroformatRenderer
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
	fmt.Println()

	dest := makeFileName(r.Title.SimpleText, f.MimeType)

	return godl.Download(f.URL, dest, true)
}

func run() error {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:\n  youtubedl [options...] <video-id>\nOptions:")
		flag.PrintDefaults()
	}
	m := flag.Bool("m", false, "Download `music` only")

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	videoID := flag.Arg(0)

	if *m {
		if err := downloadMusic(videoID); err != nil {
			return err
		}
	} else {
		if err := downloadVideo(videoID); err != nil {
			return err
		}
	}

	fmt.Println("completed!")

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
