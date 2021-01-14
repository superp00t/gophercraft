package ngdp

import (
	"fmt"
	"io"
	"net/http"

	"github.com/superp00t/gophercraft/i18n"
)

const (
	// This flag is enabled if files are retrieved from local storage using the CASC format
	OptUseCASContainer = 1 << iota
	// This flag is enabled if files are retrieved over the network (if OptUseCASContainer is not set, or if the file was not retrieved from CASC)
	OptUseTACTNetwork
)

// Agent defines the user's preferences
type Agent struct {
	Region string
	Locale i18n.Locale
	// A base URL for the NGDP entry point
	HostServer string
	// Callback for customizing how clients connect to the CDN
	DownloadFn func(url string) (io.ReadCloser, error)
}

func DefaultDownloader(url string) (io.ReadCloser, error) {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "ngdp")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	fmt.Println(response.StatusCode, request.Method, url)
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code: %s", response.Status)
	}
	return response.Body, nil
}

func DefaultAgent() *Agent {
	return &Agent{
		Locale:     i18n.English,
		DownloadFn: DefaultDownloader,
		Region:     "us",
	}
}
