package yandexmusic

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const APIRequestURL = `https://music.yandex.ru`

type API struct {
	HTTPClient *http.Client
}

type Album struct {
	Id                       interface{} `json:"id"`
	StorageDir               string      `json:"storageDir"`
	OriginalReleaseYear      int         `json:"originalReleaseYear"`
	Available                bool        `json:"available"`
	AvailableForPremiumUsers bool        `json:"availableForPremiumUsers"`
	Title                    string      `json:"title"`
	Year                     int         `json:"year"`
	Artists                  []Artist    `json:"artists"`
	CoverURI                 string      `json:"coverUri"`
	TrackCount               int         `json:"trackCount"`
	Genre                    string      `json:"genre"`
	TrackPosition            interface{} `json:"trackPosition"`
	Volumes                  []Track     `json:"volumes"`
	Lyric                    []Lyrics    `json:"lyric"`
}

func (s Album) GetID() int64 {
	switch s.Id.(type) {
	case string:
		id, err := strconv.ParseInt(s.Id.(string), 10, 32)
		if err == nil {
			return id
		}
		panic(id)
	default:
		return s.Id.(int64)
	}

}

type Artist struct {
	Id         interface{} `json:"id"`
	Cover      interface{} `json:"cover"`
	Composer   bool        `json:"composer"`
	Name       string      `json:"name"`
	Various    bool        `json:"various"`
	Decomposed interface{} `json:"decomposed"`
}

func (s Artist) GetID() int64 {
	switch s.Id.(type) {
	case string:
		id, err := strconv.ParseInt(s.Id.(string), 10, 32)
		if err == nil {
			return id
		}
		panic(id)
	default:
		return s.Id.(int64)
	}

}

type Lyrics struct {
	Id              interface{} `json:"id"`
	Lyrics          string      `json:"lyrics"`
	FullLyrics      string      `json:"fullLyrics"`
	HasRights       bool        `json:"hasRights"`
	TextLanguage    string      `json:"textLanguage"`
	ShowTranslation bool        `json:"showTranslation"`
}

func (s Lyrics) GetID() int64 {
	switch s.Id.(type) {
	case string:
		id, err := strconv.ParseInt(s.Id.(string), 10, 32)
		if err == nil {
			return id
		}
		panic(id)
	default:
		return s.Id.(int64)
	}

}

type Track struct {
	Id                       interface{}   `json:"id"`
	Albums                   []Album       `json:"albums"`
	StorageDir               string        `json:"storageDir"`
	DurationMs               int64         `json:"durationMs"`
	Title                    string        `json:"title"`
	Regions                  []interface{} `json:"regions"`
	Available                bool          `json:"available"`
	AvailableAsRbt           bool          `json:"availableAsRbt"`
	AvailableForPremiumUsers bool          `json:"availableForPremiumUsers"`
	Explicit                 bool          `json:"explicit"`
	Artists                  []Artist      `json:"artists"`
}

func (s Track) GetID() int64 {
	switch s.Id.(type) {
	case string:
		id, err := strconv.ParseInt(s.Id.(string), 10, 32)
		if err == nil {
			return id
		}
		panic(id)
	default:
		return s.Id.(int64)
	}

}

type TracksSearch struct {
	Total   int     `json:"total"`
	PerPage int     `json:"perPage"`
	Results []Track `json:"items"`
}

type ArtistsSearch struct {
	Total   int      `json:"total"`
	PerPage int      `json:"perPage"`
	Results []Artist `json:"items"`
}

type AlbumsSearch struct {
	Total   int     `json:"total"`
	PerPage int     `json:"perPage"`
	Results []Album `json:"items"`
}

type SearchResult struct {
	Playlists interface{}   `json:"playlists"`
	Albums    AlbumsSearch  `json:"albums"`
	Best      interface{}   `json:"best"`
	Artists   ArtistsSearch `json:"artists"`
	Videos    interface{}   `json:"videos"`
	Users     interface{}   `json:"videos"`
	Tracks    TracksSearch  `json:"tracks"`
	Text      string        `json:"text"`
}

type TrackResult struct {
	Counter int      `json:"counter"`
	Artists []Artist `json:"artists"`
	Aliases []string `json:"aliases"`
	Track   Track    `json:"track"`
	Lyric   []Lyrics `json:"lyric"`

	Message string `json:"message"`
}

type AlbumResult struct {
	Album
	Message string `json:"message"`
}

func (s *API) requestPOST(reqURL string, query url.Values, result interface{}) error {
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader([]byte(query.Encode())))
	if err != nil {
		return err
	}

	req.Header.Set(`Content-Type`, `application/json`)
	return s.request(req, result)
}

func (s *API) requestGET(reqURL string, result interface{}) error {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set(`Content-Type`, `application/json`)
	return s.request(req, result)
}

func (s *API) getHTTPClient() *http.Client {
	if s.HTTPClient == nil {
		s.HTTPClient = http.DefaultClient
	}
	return s.HTTPClient
}

func (s *API) request(req *http.Request, result interface{}) error {
	if resp, err := s.getHTTPClient().Do(req); err != nil {
		return err
	} else {
		if fullResp, err := ioutil.ReadAll(resp.Body); err != nil {
			return err
		} else {
			if err := json.Unmarshal(fullResp, result); err != nil {
				return err
			} else {
				return nil
			}
		}
	}
}

func (s *API) Search(text, itemType, lang string) (*SearchResult, error) {
	// itemType: all, tracks, albums, playlists, artists
	res := &SearchResult{}

	query := url.Values{}
	query.Set(`text`, text)
	query.Set(`type`, itemType)
	query.Set(`lang`, lang)
	query.Set(`external-domain`, `music.yandex.ru`)

	reqURL, _ := url.Parse(APIRequestURL)
	reqURL.Path = `/handlers/music-search.jsx`
	reqURL.RawQuery = query.Encode()

	if err := s.requestGET(reqURL.String(), res); err == nil {
		return res, nil
	} else {
		return nil, err
	}
}

func (s *API) GetTrack(albumID, trackID int64) (*TrackResult, error) {
	res := &TrackResult{}

	query := url.Values{}
	query.Set(`track`, fmt.Sprintf(`%d:%d`, trackID, albumID))

	reqURL, _ := url.Parse(APIRequestURL)
	reqURL.Path = `/handlers/track.jsx`
	reqURL.RawQuery = query.Encode()

	if err := s.requestGET(reqURL.String(), res); err == nil {
		if res.Message != `` {
			return res, errors.New(res.Message)
		}
		return res, nil
	} else {
		return nil, err
	}
}

type downloadInfo struct {
	Host string `xml:"host"`
	Path string `xml:"path"`
	TS   string `xml:"ts"`
	S    string `xml:"s"`
}

func (s *Track) GetURL() (string, error) {
	if resp, err := http.Get(`http://storage.music.yandex.ru/download-info/` + s.StorageDir + `/2.mp3`); err == nil {
		if fullResp, err := ioutil.ReadAll(resp.Body); err == nil {
			dl := &downloadInfo{}
			if err := xml.Unmarshal(fullResp, dl); err == nil {
				trackURL := url.URL{}
				trackURL.Scheme = `http`
				trackURL.Host = dl.Host
				trackURL.Path = `/get-mp3/` + GetKey(dl.Path[1:]+dl.S) + `/` + dl.TS + dl.Path
				return trackURL.String(), nil
			} else {
				return ``, err
			}
		} else {
			return ``, err
		}
	} else {
		return ``, err
	}
}

func GetKey(res string) string {
	res = `XGRlBW9FXlekgbPrRHuSiA` + strings.Replace(res, "\r\n", "\n", -1)
	return fmt.Sprintf(`%x`, md5.Sum([]byte(res)))
}

func NewAPIWithProxy(proxy string) (*API, error) {
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, err
	}
	api := &API{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		},
	}
	return api, nil
}
