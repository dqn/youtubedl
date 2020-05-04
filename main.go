package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func usage() {
	fmt.Println("Usage: youtubedl <video id>")
}

func getVideoInfo(videoID string) (*url.Values, error) {
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

	unescaped, err := url.QueryUnescape(string(b))
	if err != nil {
		return nil, err
	}

	v, err := url.ParseQuery(unescaped)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func run() error {
	if len(os.Args) != 2 {
		usage()
		os.Exit(2)
	}

	v, err := getVideoInfo(os.Args[1])
	if err != nil {
		return err
	}
	fmt.Println(*v)

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
