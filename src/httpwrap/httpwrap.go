// Wrapper metodi "net/http" per facilitare lo scraping
package httpwrap

import (
	"bytes"
	"net/http"
	"net/url"
)

func Get(url string, header http.Header, params url.Values, cookies []*http.Cookie) (*http.Response, error) {
	request, _ := http.NewRequest("GET", url+params.Encode(), nil)
	request.Header = header

	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func Post(url string, header http.Header, params *bytes.Buffer, cookies []*http.Cookie) (*http.Response, error) {
	request, err := http.NewRequest("POST", url, params)
	request.Header = header

	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}
