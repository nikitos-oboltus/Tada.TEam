package tdclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/tada-team/tdproto"
)

type Session struct {
	Timeout  time.Duration
	logger   *log.Logger
	server   url.URL
	token    string
	cookie   string
	features *tdproto.Features
}

func NewSession(server string) (Session, error) {
	s := Session{
		Timeout: 10 * time.Second,
	}

	s.logger = log.New(os.Stdout, "tdclient: ", log.LstdFlags|log.Lmicroseconds|log.Lmsgprefix)
	s.SetVerbose(false)

	u, err := url.Parse(server)
	if err != nil {
		return Session{}, err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return Session{}, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
	s.server = *u

	return s, nil
}

func (s *Session) Features() (*tdproto.Features, error) {
	if s.features == nil {
		if err := s.doGet("/features.json", &s.features); err != nil {
			return s.features, err
		}
	}
	return s.features, nil
}

func (s *Session) SetToken(v string) {
	s.token = v
}

func (s *Session) SetCookie(v string) {
	s.cookie = v
}

func (s *Session) SetVerbose(v bool) {
	if v {
		s.logger.SetOutput(os.Stdout)
	} else {
		s.logger.SetOutput(ioutil.Discard)
	}
}

func (s Session) httpClient() *http.Client {
	return &http.Client{
		Timeout: s.Timeout,
		//Transport: &http.Transport{
		//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		//},
	}
}

func (s Session) url(path string) string {
	s.server.Path = path
	return s.server.String()
}

func (s Session) doGet(path string, resp interface{}) error {
	return s.doRaw("GET", path, nil, resp)
}

func (s Session) doPost(path string, data, v interface{}) error {
	return s.doRaw("POST", path, data, v)
}

func (s Session) doDelete(path string, resp interface{}) error {
	return s.doRaw("DELETE", path, nil, resp)
}

func (s Session) doRaw(method, path string, data, v interface{}) error {
	client := s.httpClient()

	path = s.url(path)

	var buf *bytes.Buffer
	if data == nil {
		s.logger.Println(method, path)
		buf = bytes.NewBuffer([]byte{})
	} else {
		s.logger.Println(method, path, debugJSON(data))
		b, err := json.Marshal(data)
		if err != nil {
			return errors.Wrap(err, "json marshal fail")
		}
		buf = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, path, buf)
	if err != nil {
		return errors.Wrap(err, "new request fail")
	}

	if s.token != "" {
		req.Header.Set("token", s.token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "client do fail")
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read body fail")
	}

	if err := JSON.Unmarshal(respData, &v); err != nil {
		return errors.Wrapf(err, "unmarshal fail on: %s", string(respData))
	}

	s.logger.Println(debugJSON(v))

	return nil
}
