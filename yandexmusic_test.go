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

func TestAPI_GetAlbum(t *testing.T) {
	api, err := NewAPIWithProxy(os.Getenv(`proxy`))
	if err != nil {
		t.Fatal(err)
	}
	if res, err := api.GetAlbum(1, 0); err != nil {
		t.Fatal(err)
	} else {
		t.Log(res)
	}
	if res, err := api.GetAlbum(0, 1); err != nil {
		t.Fatal(err)
	} else {
		t.Log(res)
	}

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

func TestAPI_GetArtist(t *testing.T) {
	api, err := NewAPIWithProxy(os.Getenv(`proxy`))
	if err != nil {
		t.Fatal(err)
	}
	res, err := api.GetArtist(711232, ``)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
	res, err = api.GetArtist(711232, `tracks`)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}
