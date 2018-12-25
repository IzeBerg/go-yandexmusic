package yandexmusic

import (
	"os"
	"testing"
)

func TestSearch(t *testing.T) {
	api, err := NewAPI(Config{
		Device: Device{
			DeviceID:`377c5ae26b09fccd72deae0a95425559`,
			UUID:`3cfccdaf75dcf98b917a54afe50447ba`,
			PackageID: DefaultPackageName,
			ClientID: `23cabbbdc6cd418abb4b39c32c41195d`,
			ClientSecret: `53bc75238f0c4d08a118e51fe9203300`,
		},
		User: User{
			Username:os.Getenv(`username`),
			Password:os.Getenv(`password`),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := api.Search(`M83 - Go!`, `all`, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	track := res.Result.Tracks.Results[0]
	trackURL, err := track.GetURL()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(track.Artists[0].Name, `-`, track.Title, trackURL)
}

