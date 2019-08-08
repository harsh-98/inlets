package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
	"context"
	"os"
	"strings"

	"github.com/harsh-98/inlets/pkg/router"
	"github.com/harsh-98/inlets/pkg/transport"
	"github.com/rancher/remotedialer"
	"github.com/twinj/uuid"
	"k8s.io/apimachinery/pkg/util/proxy"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// Server for the exit-node of inlets
type Server struct {
	Port   int
	router router.Router
	server *remotedialer.Server
	mClient *mongo.Client
	DisableWrapTransport bool
}
func mongoClient()*mongo.Client{
	if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
	}
	mongoUrl := getMongoUrl()
	client, _ := mongo.NewClient(options.Client().ApplyURI(mongoUrl))

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_ = client.Connect(ctx)

	return client
}
func (s *Server)getToken(token string, t *bson.M) {
	collection := s.mClient.Database("test").Collection("tokens")
	filter := bson.M{"token": token}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, filter).Decode(&t)
	if err != nil {
		log.Fatal(err)
	}
}
func getMongoUrl() string{
	MONGOURL, exists := os.LookupEnv("MONGOURL")

    if exists {
	return MONGOURL
	}
	return ""
}

// Serve traffic
func (s *Server) Serve() {
	s.mClient = mongoClient()
	s.server = remotedialer.New(s.authorized, remotedialer.DefaultErrorWriter)
	s.router.Server = s.server

	http.HandleFunc("/", s.proxy)
	http.HandleFunc("/tunnel", s.tunnel)

	log.Printf("Listening on :%d\n", s.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) tunnel(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
	s.router.Remove(r)
}

func (s *Server) proxy(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Host)
	route := s.router.Lookup(r)
	if route == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	inletsID := uuid.Formatter(uuid.NewV4(), uuid.FormatHex)
	log.Printf("[%s] proxy %s %s %s", inletsID, r.Host, r.Method, r.URL.String())
	r.Header.Set(transport.InletsHeader, inletsID)

	u := *r.URL
	u.Host = r.Host
	u.Scheme = route.Scheme

	httpProxy := proxy.NewUpgradeAwareHandler(&u, route.Transport, !s.DisableWrapTransport, false, s)
	httpProxy.ServeHTTP(w, r)
}

func (s Server) Error(w http.ResponseWriter, req *http.Request, err error) {
	remotedialer.DefaultErrorWriter(w, req, http.StatusInternalServerError, err)
}

func (s *Server) dialerFor(id, host string) remotedialer.Dialer {
	return func(network, address string) (net.Conn, error) {
		return s.server.Dial(id, time.Minute, network, host)
	}
}

func (s *Server) tokenValid(req *http.Request) bool {
	auth := req.Header.Get("Authorization")
	var t bson.M
	token:=strings.Split(auth, " ")[1]
	fmt.Println(token)
	s.getToken(token, &t)
	fmt.Println(t["revoked"] != true)
	return t["revoked"] != true
}

func (s *Server) authorized(req *http.Request) (id string, ok bool, err error) {
	defer func() {
		if id == "" {
			// empty id is also an auth failure
			ok = false
		}
		if !ok || err != nil {
			// don't let non-authed request clear routes
			req.Header.Del(transport.InletsHeader)
		}
	}()

	if !s.tokenValid(req) {
		return "", false, nil
	}

	return s.router.Add(req), true, nil
}
