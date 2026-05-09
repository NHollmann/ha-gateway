package hagateway

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

func addCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With")
		handler.ServeHTTP(w, r)
	})
}

type Gateway interface {
	http.Handler
	Ping() error
}

type gateway struct {
	baseHandler http.Handler
	clients     Clients
	serverURL   *url.URL
	authToken   string
}

func New(remoteUrl string, authToken string, clients Clients) Gateway {
	serverUrl, err := url.Parse(remoteUrl)
	if err != nil {
		log.Fatal("URL failed to parse")
	}

	return &gateway{
		clients: clients,
		baseHandler: addCORS(&httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				pr.SetURL(serverUrl)
				pr.Out.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
			},
		}),
		serverURL: serverUrl,
		authToken: authToken,
	}
}

const MSG_FORBIDDEN = "Forbidden"
const PATH_PREFIX = "/api/states/"
const AUTHORIZATION_PREFIX = "Bearer hagakey_"

func (g *gateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.EscapedPath()
	if !strings.HasPrefix(path, PATH_PREFIX) {
		log.Printf("Error: Forbidden (%s): Invalid path prefix", req.RemoteAddr)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(MSG_FORBIDDEN))
		return
	}

	authHeader := req.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, AUTHORIZATION_PREFIX) {
		log.Printf("Error: Forbidden (%s): Missing authorization header", req.RemoteAddr)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(MSG_FORBIDDEN))
		return
	}

	authToken := strings.TrimPrefix(authHeader, AUTHORIZATION_PREFIX)
	client := g.clients.FindByToken(authToken)
	if client == nil {
		log.Printf("Error: Forbidden (%s): Found no client with matching api key", req.RemoteAddr)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(MSG_FORBIDDEN))
		return
	}

	entity := strings.TrimPrefix(path, PATH_PREFIX)
	switch req.Method {
	case "GET":
		if client.CanReadEntity(entity) {
			log.Printf("GET (%s) '%s'", client.Name, entity)
			g.baseHandler.ServeHTTP(w, req)
			return
		} else {
			log.Printf("Error: Forbidden (%s): GET not allowed on entity '%s'", client.Name, entity)
		}
	case "POST":
		if client.CanWriteEntity(entity) {
			log.Printf("POST (%s) '%s'", client.Name, entity)
			g.baseHandler.ServeHTTP(w, req)
			return
		} else {
			log.Printf("Error: Forbidden (%s): POST not allowed on entity '%s'", client.Name, entity)
		}
	}

	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(MSG_FORBIDDEN))
}

func (g *gateway) Ping() error {
	if g.serverURL == nil {
		return fmt.Errorf("no server URL configured")
	}

	pingURL := *g.serverURL
	pingURL.Path = "/api/"

	req, err := http.NewRequest("GET", pingURL.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.authToken))

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusMultiStatus {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
