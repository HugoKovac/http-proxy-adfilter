package proxy

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/proxy"
)

func GetDomainNotFound(t *testing.T, requestURL string, proxyURL string) {
	_, originalError, proxyResp, proxyError := GetOriginAndProxy(t, requestURL, proxyURL)

	if originalError == nil {
		t.Logf("%s could be found", requestURL)
		t.Fail()
	} else if !strings.Contains(originalError.Error(), "no such host") {
		t.Logf("Other error than not resolved: (%s)", originalError.Error())
		t.Fail()
	} else if proxyError != nil {
		t.Logf("Error request by proxy: %s", proxyError.Error())
		t.Fail()
	} else {
		defer proxyResp.Body.Close()

		var body []byte
		body, err := io.ReadAll(proxyResp.Body)
		if err != nil {
			t.Log("Error reading proxied request body")
			t.Fail()
		} else if !strings.Contains(string(body), "no such host") {
			t.Logf("Expecting: '%s' got '%s'", "no such host", body)
			t.Fail()
		} else if proxyResp.StatusCode != 502 {
			t.Logf("Expecting status code 502 Bad Gateway go %s", proxyResp.Status)
			t.Fail()
		}
	}

}

func HeaderDiff(t *testing.T, originalH http.Header, proxyH http.Header) {
	for key := range originalH {
		if strings.Compare(originalH.Get(key), proxyH.Get(key)) != 0 {
			t.Logf("original[%s] = %s | proxy[%s] = %s", key, originalH.Get(key), key, proxyH.Get(key))
			proxyH.Del(key)
		}
	}

	for key := range proxyH {
		t.Log(key, proxyH.Get(key))
	}
}

func GetOriginAndProxy(t *testing.T, requestURL string, proxyURL string) (normalResp *http.Response, originalError error, proxyResp *http.Response, proxyError error) {
	url, originalError := url.Parse("http://" + proxy.HOST + ":" + proxy.PORT)
	if originalError != nil {
		t.Fatal(proxyURL, "not valid as URL")
	}
	proxy := http.ProxyURL(url)
	transport := &http.Transport{Proxy: proxy}
	client := &http.Client{}
	proxyClient := &http.Client{Transport: transport}

	normalResp, originalError = client.Get(requestURL)
	proxyResp, proxyError = proxyClient.Get(requestURL)

	return normalResp, originalError, proxyResp, proxyError
}

func GetCompare(t *testing.T, requestURL string, proxyURL string) {
	normalResp, originalError, proxyResp, proxyError := GetOriginAndProxy(t, requestURL, proxyURL)

	if originalError != nil || proxyError != nil {
		if originalError != nil && proxyError != nil && strings.Compare(originalError.Error(), proxyError.Error()) != 0 {
			t.Errorf("Got '%s' when expecting '%s'", proxyError, originalError)
			t.Fail()
			return
		} else if originalError != nil && proxyError == nil {
			t.Error("Our's didn't encounter an error when the original does\noriginalError: ", originalError)
			t.Fail()
			return
		} else if originalError == nil && proxyError != nil {
			t.Error("The original didn't encounter an error when ours does\nproxyError: ", proxyError)
			t.Fail()
			return
		} else {
			t.Logf("Both returning the same error '%s' and '%s'", originalError, proxyError)
		}
	}

	if normalResp.StatusCode != proxyResp.StatusCode {
		t.Errorf("Got status code %d when expecting %d", proxyResp.StatusCode, normalResp.StatusCode)
		t.Fail()
		return
	}

	defer normalResp.Body.Close()
	defer proxyResp.Body.Close()

	var normalBody []byte
	var proxyBody []byte

	_, originalError = normalResp.Body.Read(normalBody)
	_, proxyError = proxyResp.Body.Read(proxyBody)

	if originalError != nil || proxyError != nil {
		t.Log("Error reading body")
	}

	if normalBody != nil && proxyBody != nil {

		for i := range normalBody {
			if normalBody[i] != proxyBody[i] {
				t.Errorf("Got body %s when expecting %s", proxyBody, normalBody)
				t.Fail()
				return
			}
		}
	}

	// HeaderDiff(t, normalResp.Header, proxyResp.Header)

	t.Log("No diff")
}

func TestSimpleRequest(t *testing.T) {
	GetCompare(t, "http://www.google.com", "http://" + proxy.HOST + ":" + proxy.PORT)
	GetCompare(t, "http://www.google.com/index.html", "http://" + proxy.HOST + ":" + proxy.PORT)
	GetDomainNotFound(t, "http://www.google.comm", "http://" + proxy.HOST + ":" + proxy.PORT)
}
