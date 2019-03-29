package yandexmusic

import (
	"os"
	"testing"
)

func TestAPI_GetTrack(t *testing.T) {
	api, err := NewAPIWithProxy(os.Getenv(`proxy`))
	if err != nil {
		t.Fatal(err)
	}
	res, err := api.GetTrack(3542, 43117)
	if err != nil {
		t.Fatal(err)
	}
	trackURL, err := res.Track.GetURL()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res.Track.Artists[0].Name, `-`, res.Track.Title, trackURL)
}

func TestAPI_Search(t *testing.T) {
	api, err := NewAPIWithProxy(os.Getenv(`proxy`))
	if err != nil {
		t.Fatal(err)
	}
	res, err := api.Search(`M83 - Go!`, `all`, `en`)
	if err != nil {
		t.Fatal(err)
	}
	track := res.Tracks.Results[0]
	trackURL, err := track.GetURL()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(track.Artists[0].Name, `-`, track.Title, trackURL)
}
