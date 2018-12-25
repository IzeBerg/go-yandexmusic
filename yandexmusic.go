package yandexmusic

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const DefaultPackageName = `ru.yandex.music`
const OAuthURL = `https://oauth.mobile.yandex.net/1/token`
const APIRequestURL = `https://api.music.yandex.net`

type OAuth struct {
	DeviceID    string
	UUID        string
	PackageName string
}

type Device struct {
	DeviceID  string
	UUID      string
	PackageID string

	ClientID     string
	ClientSecret string
}

type User struct {
	Username    string
	Password    string
	AccessToken string
	UID         int64
	ExpiresIn   int64
}

type Config struct {
	Device Device
	User   User
}

type API struct {
	Config Config
}

type TokenResult struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	TokenType        string `json:"token_type"`
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	UID              int64  `json:"uid"`
}

type Album struct {
	Id                       int         `json:"id"`
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
}

type Artist struct {
	Id         int         `json:"id"`
	Cover      interface{} `json:"cover"`
	Composer   bool        `json:"composer"`
	Name       string      `json:"name"`
	Various    bool        `json:"various"`
	Decomposed interface{} `json:"decomposed"`
}

type Track struct {
	Id                       int           `json:"id"`
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

type TracksSearchResult struct {
	Total   int     `json:"total"`
	PerPage int     `json:"perPage"`
	Results []Track `json:"results"`
}

type ArtistsSearchResult struct {
	Total   int      `json:"total"`
	PerPage int      `json:"perPage"`
	Results []Artist `json:"results"`
}

type AlbumsSearchResult struct {
	Total   int     `json:"total"`
	PerPage int     `json:"perPage"`
	Results []Album `json:"results"`
}

type SearchResult struct {
	Playlists  interface{}         `json:"playlists"`
	Albums     AlbumsSearchResult  `json:"albums"`
	Best       interface{}         `json:"best"`
	Artists    ArtistsSearchResult `json:"artists"`
	Videos     interface{}         `json:"videos"`
	Tracks     TracksSearchResult  `json:"tracks"`
	NonCorrect bool                `json:"noncorrect"`
	Text       string              `json:"text"`
}

type APISearchResult struct {
	InvocationInfo interface{}  `json:"invocationInfo"`
	Result         SearchResult `json:"result"`
}

func NewAPI(config Config) (*API, error) {
	api := &API{Config: config}
	return api, api.Auth()
}

func (s *API) requestPOST(reqURL string, query url.Values, result interface{}, headers map[string]string) error {
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader([]byte(query.Encode())))
	if err != nil {
		return err
	}

	req.Header.Set(`Content-Type`, `application/json`)

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	return s.request(req, result)
}

func (s *API) requestGET(reqURL string, result interface{}, headers map[string]string) error {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set(`Content-Type`, `application/json`)

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	return s.request(req, result)
}

func (s *API) request(req *http.Request, result interface{}) error {
	if resp, err := http.DefaultClient.Do(req); err != nil {
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

func (s *API) authGetToken() (*TokenResult, error) {
	query := url.Values{}
	query.Add(`grant_type`, `password`)
	query.Add(`username`, s.Config.User.Username)
	query.Add(`password`, s.Config.User.Password)
	query.Add(`client_id`, s.Config.Device.ClientID)
	query.Add(`client_secret`, s.Config.Device.ClientSecret)

	res := &TokenResult{}
	if err := s.requestPOST(OAuthURL, query, res, nil); err != nil {
		return nil, err
	} else {
		if res.Error != `` {
			panic(res)
		}
		return res, nil
	}

}

func (s *API) authSetupToken() error {
	query := url.Values{}
	query.Add(`grant_type`, `x-token`)
	query.Add(`access_token`, s.Config.User.AccessToken)
	query.Add(`client_id`, s.Config.Device.ClientID)
	query.Add(`client_secret`, s.Config.Device.ClientSecret)
	query.Add(`device_id`, s.Config.Device.DeviceID)
	query.Add(`uuid`, s.Config.Device.UUID)
	query.Add(`package_name`, s.Config.Device.PackageID)

	res := &TokenResult{}
	if err := s.requestPOST(OAuthURL, query, res, nil); err != nil {
		return err
	} else {
		if res.Error != `` {
			panic(res)
		}
		s.Config.User.AccessToken = res.AccessToken
		s.Config.User.UID = res.UID
		s.Config.User.ExpiresIn = res.ExpiresIn
		return nil
	}
}

func (s *API) Auth() error {
	if token, err := s.authGetToken(); err != nil {
		return s.authSetupToken()
	} else {
		s.Config.User.AccessToken = token.AccessToken
		s.Config.User.UID = token.UID
		s.Config.User.ExpiresIn = token.ExpiresIn
	}
	return nil
}

func (s *API) getAuthHeaders() map[string]string {
	return map[string]string{
		`Authorization`: `OAuth ` + s.Config.User.AccessToken,
	}
}

func (s *API) Search(text string, stype string, page int, noncorrect bool) (*APISearchResult, error) {
	res := &APISearchResult{}

	query := url.Values{}
	query.Set(`text`, text)
	query.Set(`type`, stype)
	query.Set(`page`, strconv.Itoa(page))
	query.Set(`noncorrect`, strconv.FormatBool(noncorrect))

	reqURL, _ := url.Parse(APIRequestURL)
	reqURL.Path = `/search`
	reqURL.RawQuery = query.Encode()

	if err := s.requestGET(reqURL.String(), res, s.getAuthHeaders()); err == nil {
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
