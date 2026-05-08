package hagateway

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
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
}

type gateway struct {
	baseHandler http.Handler
	clients     Clients
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
	}
}

const MSG_FORBIDDEN = "Forbidden"
const PATH_PREFIX = "/api/states/"
const AUTHORIZATION_PREFIX = "Bearer hagakey_"

func (g *gateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.EscapedPath()
	if !strings.HasPrefix(path, PATH_PREFIX) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(MSG_FORBIDDEN))
		return
	}

	authHeader := req.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, AUTHORIZATION_PREFIX) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(MSG_FORBIDDEN))
		return
	}

	authToken := strings.TrimPrefix(authHeader, AUTHORIZATION_PREFIX)
	client := g.clients.FindByToken(authToken)
	if client == nil {
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
		}
	case "POST":
		if client.CanWriteEntity(entity) {
			log.Printf("POST (%s) '%s'", client.Name, entity)
			g.baseHandler.ServeHTTP(w, req)
			return
		}
	}

	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(MSG_FORBIDDEN))
}
