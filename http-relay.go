package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	log         = logrus.New()
	env         config
	proxyClient *http.Client
)

type config struct {
	Port              int    `yaml:"port"`
	TargetURL         string `yaml:"targetUrl"`
	ConnectionTimeout int    `yaml:"connectionTimeout"`
	Proxy             struct {
		URL      string `yaml:"url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"proxy"`
}

type requestError struct {
	Error string `json:"error"`
}

func main() {
	log.SetFormatter(&logrus.JSONFormatter{})
	loadConfig("./config.yml")
	initProxyClient()

	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", env.Port), nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req, err := http.NewRequest(
		r.Method,
		fmt.Sprintf("%s%s", env.TargetURL, r.RequestURI),
		io.Reader(r.Body),
	)

	if err != nil {
		log.Error(err)
		w.Write(newErrorResponse(errors.New("Failed to build a request")))
		return
	}

	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Accept", "application/json")

	log.Println("Forwarding to", env.TargetURL)
	resp, err := proxyClient.Do(req)

	if err != nil {
		log.Error(err)
		w.Write(newErrorResponse(errors.New("Request failed")))
		return
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Error(err)
		w.Write(newErrorResponse(errors.New("Failed to read response body")))
		return
	}

	_, err = w.Write(b)
}

func newErrorResponse(err error) []byte {
	re := requestError{err.Error()}
	e, err := json.Marshal(re)

	if err != nil {
		log.Error(err)
	}

	return e
}

func loadConfig(path string) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(contents, &env)
	if err != nil {
		log.Fatal(err)
	}
}

func proxyURL() *url.URL {
	proxyRawURL := env.Proxy.URL

	// add optional auth for proxy
	if len(env.Proxy.Username) > 0 && len(env.Proxy.Password) > 0 {
		sl := strings.Split(proxyRawURL, "://")
		sl[1] = fmt.Sprintf("%s:%s@%s",
			env.Proxy.Username,
			env.Proxy.Password,
			sl[1],
		)

		proxyRawURL = strings.Join(sl, "://")
	}

	u, err := url.Parse(proxyRawURL)

	if err != nil {
		log.Fatal(err)
	}

	return u
}

func initProxyClient() {
	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyURL()),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	timeout := time.Duration(env.ConnectionTimeout) * time.Second

	proxyClient = &http.Client{Transport: tr, Timeout: timeout}
}
