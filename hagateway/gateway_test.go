package hagateway_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NHollmann/ha-gateway/hagateway"
)

const BACKEND_RESPONSE = "TEST"

type testServer struct {
	destination *httptest.Server
	frontend    *httptest.Server
}

func (ts *testServer) Close() {
	ts.destination.Close()
	ts.frontend.Close()
}

func (ts *testServer) FrontendURL() string {
	return ts.frontend.URL
}

func (ts *testServer) Client() *http.Client {
	return ts.frontend.Client()
}

func newTestServer(t *testing.T) *testServer {
	destinationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer AUTH" {
			t.Errorf("Wrong authorization header: %s", r.Header.Get("Authorization"))
		}
		w.Write([]byte(BACKEND_RESPONSE))
	}))

	clients := hagateway.Clients{}
	clients.Add(&hagateway.Client{
		Name:      "Wildcard",
		TokenHash: "94ee059335e587e501cc4bf90613e0814f00a7b08bc7c648fd865a2af6a22cc2",
		CanWrite:  true,
		Entities:  []hagateway.Entity{{Name: "*"}},
	})
	clients.Add(&hagateway.Client{
		Name:      "Limited",
		TokenHash: "bb7edb19fb19a0a455efb2c4d54957b394d8bcf246b478e51ceb2cb44328447f",
		CanWrite:  false,
		Entities:  []hagateway.Entity{{Name: "binary_sensor.light_switch"}},
	})

	proxyHandler := hagateway.New(destinationServer.URL, "AUTH", clients)
	frontend := httptest.NewServer(proxyHandler)

	return &testServer{
		destination: destinationServer,
		frontend:    frontend,
	}
}

func readBody(res *http.Response) string {
	bodyBytes, _ := io.ReadAll(res.Body)
	res.Body.Close()
	return string(bodyBytes)
}

func TestProxy(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	getReq, _ := http.NewRequest("GET", server.FrontendURL()+"/api/states/", nil)
	getReq.Header.Add("Authorization", "Bearer TEST")
	res, err := server.Client().Do(getReq)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Wrong status code: %d, expected %d", res.StatusCode, http.StatusOK)
	}
	if g, e := readBody(res), BACKEND_RESPONSE; g != e {
		t.Errorf("got body %q; expected %q", g, e)
	}

	postReq, _ := http.NewRequest("POST", server.FrontendURL()+"/api/states/", nil)
	postReq.Header.Add("Authorization", "Bearer TEST")
	res, err = server.Client().Do(postReq)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Wrong status code: %d, expected %d", res.StatusCode, http.StatusOK)
	}
	if g, e := readBody(res), BACKEND_RESPONSE; g != e {
		t.Errorf("got body %q; expected %q", g, e)
	}
}

func TestForbiddenPaths(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	getReq, _ := http.NewRequest("GET", server.FrontendURL()+"/bad-url", nil)
	getReq.Header.Add("Authorization", "Bearer TEST")
	res, err := server.Client().Do(getReq)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("Wrong status code: %d, expected %d", res.StatusCode, http.StatusForbidden)
	}

	postReq, _ := http.NewRequest("POST", server.FrontendURL(), nil)
	postReq.Header.Add("Authorization", "Bearer TEST")
	res, err = server.Client().Do(postReq)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("Wrong status code: %d, expected %d", res.StatusCode, http.StatusForbidden)
	}
}

func TestWrongAuth(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	getReq, _ := http.NewRequest("GET", server.FrontendURL()+"/api/states/abc", nil)
	res, err := server.Client().Do(getReq)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("Wrong status code: %d, expected %d", res.StatusCode, http.StatusForbidden)
	}

	getReq, _ = http.NewRequest("GET", server.FrontendURL()+"/api/states/abc", nil)
	getReq.Header.Add("Authorization", "Bearer ABC")
	res, err = server.Client().Do(getReq)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("Wrong status code: %d, expected %d", res.StatusCode, http.StatusForbidden)
	}

	postReq, _ := http.NewRequest("POST", server.FrontendURL()+"/api/states/binary_sensor.light_switch", nil)
	postReq.Header.Add("Authorization", "Bearer HomeAssistant")
	res, err = server.Client().Do(postReq)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("Wrong status code: %d, expected %d", res.StatusCode, http.StatusForbidden)
	}
}
