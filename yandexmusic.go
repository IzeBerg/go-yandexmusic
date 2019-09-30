package yandexmusic

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const APIRequestURL = `https://music.yandex.ru`

var ErrNotFound = errors.New(`not found`)

func GetID(id interface{}) int64 {
	switch id.(type) {
	case string:
		if id, err := strconv.ParseInt(id.(string), 10, 32); err == nil {
			return id
		} else {
			panic(err)
		}
	case float64:
		return int64(id.(float64))
	case int64:
		return int64(id.(float64))
	default:
		panic(fmt.Errorf(`unknown type of id: %s`, id))
	}
}

type API struct {
	HTTPClient *http.Client
}

type ErrorContainer struct {
	Message string `json:"message"`
}

func (s ErrorContainer) Error() string {
	return s.Message
}

type Album struct {
	ErrorContainer

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
	Volumes                  [][]Track   `json:"volumes"`
	Lyric                    []Lyrics    `json:"lyric"`
}

func (s Album) GetID() int64 {
	return GetID(s.Id)
}

type Artist struct {
	ErrorContainer

	Id         interface{} `json:"id"`
	Cover      interface{} `json:"cover"`
	Composer   bool        `json:"composer"`
	Name       string      `json:"name"`
	Various    bool        `json:"various"`
	Decomposed interface{} `json:"decomposed"`
}

func (s Artist) GetID() int64 {
	return GetID(s.Id)
}

type Lyrics struct {
	ErrorContainer

	Id              interface{} `json:"id"`
	Lyrics          string      `json:"lyrics"`
	FullLyrics      string      `json:"fullLyrics"`
	HasRights       bool        `json:"hasRights"`
	TextLanguage    string      `json:"textLanguage"`
	ShowTranslation bool        `json:"showTranslation"`
}

func (s Lyrics) GetID() int64 {
	return GetID(s.Id)
}

type Track struct {
	ErrorContainer

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
	return GetID(s.Id)
}

type TracksSearch struct {
	ErrorContainer

	Total   int     `json:"total"`
	PerPage int     `json:"perPage"`
	Results []Track `json:"items"`
}

type ArtistsSearch struct {
	ErrorContainer

	Total   int      `json:"total"`
	PerPage int      `json:"perPage"`
	Results []Artist `json:"items"`
}

type AlbumsSearch struct {
	ErrorContainer

	Total   int     `json:"total"`
	PerPage int     `json:"perPage"`
	Results []Album `json:"items"`
}

type SearchResult struct {
	ErrorContainer

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
	ErrorContainer

	Counter       int      `json:"counter"`
	Artists       []Artist `json:"artists"`
	Aliases       []string `json:"aliases"`
	Track         Track    `json:"track"`
	SimilarTracks []Track  `json:"similarTracks"`
	Lyric         []Lyrics `json:"lyric"`
}

type ArtistResult struct {
	ErrorContainer
	Artist        Artist        `json:"artist"`
	Similar       []Artist      `json:"similar"`
	AllSimilar    []Artist      `json:"allSimilar"`
	Albums        []Album       `json:"albums"`
	AlsoAlbums    []Album       `json:"alsoAlbums"`
	Tracks        []Track       `json:"tracks"`
	TrackIds      []string      `json:"trackIds"`
	Playlists     []interface{} `json:"playlists"`
	PlaylistIDs   []interface{} `json:"playlistIds"`
	HasPromotions bool          `json:"hasPromotions"`
	LikesCount    int           `json:"likesCount"`
	Redirected    bool          `json:"redirected"`
	Radio         struct {
		Available bool `json:"available"`
	} `json:"radio"`
}

func (s *ArtistResult) GetTrackIds() []int64 {
	var ids []int64
	for _, src := range s.TrackIds {
		if id, err := strconv.ParseInt(src, 10, 64); err == nil {
			ids = append(ids, id)
		} else {
			panic(err)
		}
	}
	return ids
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
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			return ErrNotFound
		} else if fullResp, err := ioutil.ReadAll(resp.Body); err != nil {
			return err
		} else {
			return json.Unmarshal(fullResp, result)
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
		if res.Message != `` {
			return nil, res
		}
		return res, nil
	} else {
		return nil, err
	}
}

func (s *API) GetAlbum(albumID, byTrack int64) (*Album, error) {
	res := &Album{}

	query := url.Values{}
	if albumID != 0 {
		query.Set(`album`, fmt.Sprintf(`%d`, albumID))
	}
	if byTrack != 0 {
		query.Set(`byTrack`, fmt.Sprintf(`%d`, byTrack))
	}

	reqURL, _ := url.Parse(APIRequestURL)
	reqURL.Path = `/handlers/album.jsx`
	reqURL.RawQuery = query.Encode()

	if err := s.requestGET(reqURL.String(), res); err == nil {
		if res.Message != `` {
			return nil, res
		}
		return res, nil
	} else {
		return nil, err
	}
}

func (s *API) GetTrack(albumID, trackID int64) (*TrackResult, error) {
	if albumID == 0 {
		if album, err := s.GetAlbum(0, trackID); err == nil {
			albumID = album.GetID()
		} else {
			return nil, err
		}
	}

	res := &TrackResult{}

	query := url.Values{}
	query.Set(`track`, fmt.Sprintf(`%d:%d`, trackID, albumID))

	reqURL, _ := url.Parse(APIRequestURL)
	reqURL.Path = `/handlers/track.jsx`
	reqURL.RawQuery = query.Encode()

	if err := s.requestGET(reqURL.String(), res); err == nil {
		if res.Message != `` {
			return nil, res
		}
		return res, nil
	} else {
		return nil, err
	}
}

func (s *API) GetArtist(artistID int64) (*ArtistResult, error) {
	res := &ArtistResult{}

	query := url.Values{}
	query.Set(`artist`, strconv.FormatInt(artistID, 10))

	reqURL, _ := url.Parse(APIRequestURL)
	reqURL.Path = `/handlers/artist.jsx`
	reqURL.RawQuery = query.Encode()

	if err := s.requestGET(reqURL.String(), res); err == nil {
		if res.Message != `` {
			return nil, res
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
	if resp, err := http.Get(`http://storage.mds.yandex.net/download-info/` + s.StorageDir + `/2`); err == nil {
		defer resp.Body.Close()
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
	var proxyURL *url.URL
	if proxy != `` {
		if _url, err := url.Parse(proxy); err != nil {
			return nil, err
		} else {
			proxyURL = _url
		}
	}
	client := &http.Client{}
	if proxyURL != nil {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}
	api := &API{HTTPClient: client}
	return api, nil
}
