package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"./data"
	"./model"

	_ "github.com/go-sql-driver/mysql"
)

var dbh *sql.DB
var API_KEY string

const longForm = "2006-01-02T15:04:05Z0700"

func init() {

	API_KEY = os.Getenv("API_KEY")
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = os.Getenv("USER")
	}

	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_DATABASE")
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	db, err := sql.Open("mysql", dbUser+":"+dbPass+"@tcp("+dbHost+")/"+dbName+"?collation=latin1_swedish_ci")
	if err != nil {
		fmt.Println(dbUser + ":XXX" + "@tcp(" + dbHost + ")/" + dbName)
		panic(err.Error())
	}

	// Open doesn't open a connection, se let's Ping() our db
	err = db.Ping()
	if err != nil {
		fmt.Println(dbUser + ":XXX" + "@tcp(" + dbHost + ")/" + dbName)
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	dbh = db
}

type Server struct {
	logger *log.Logger
	mux    *http.ServeMux
}

func NewServer(options ...func(*Server)) *Server {
	s := &Server{
		logger: log.New(os.Stdout, "", 0),
		mux:    http.NewServeMux(),
	}

	for _, f := range options {
		f(s)
	}

	s.mux.HandleFunc("/", s.api)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type postData struct {
	APIKey      string `json:"api_key"`
	LastUpdated string `json:"last_updated"`
}

func (s *Server) api(w http.ResponseWriter, r *http.Request) {
	cType := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.Index(cType, "application/json") != 0 {
		bailOut(w, fmt.Sprintf("invalid content-type: %s", cType))
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	t := time.Now()
	s.logger.Println(ip + "\t" + t.Format("2006-01-02 15:04:05") +
		"\t" + r.Header.Get("User-Agent"))
	//s.logger.Println("Remote address: ", r.RemoteAddr)
	if r.Method != "POST" {
		bailOut(w, "invalid request")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	var pData postData
	if err := decoder.Decode(&pData); err != nil {
		s.logger.Println("invalid request: " + err.Error())
		bailOut(w, fmt.Sprintf("invalid request: %s", err.Error()))
		return
	}

	if pData.APIKey != API_KEY {
		fmt.Printf("** got invalid api_key [%s]\n", pData.APIKey)
		bailOut(w, "invalid api_key")
		return
	}

	itemGetter := data.Items{DB: dbh}
	var items []model.Item
	var err error
	if pData.LastUpdated == "" {
		items, err = itemGetter.GetAll()
	} else {

		var tLastUpdated time.Time
		tLastUpdated, err = time.Parse(longForm, pData.LastUpdated)
		if err != nil {
			bailOut(w, fmt.Sprintf("invalid time: %s", err.Error()))
			return
		}

		items, err = itemGetter.GetAll(tLastUpdated.Format("2006-01-02 15:04:05"))
	}

	output := struct {
		Status   string       `json:"status"`
		Messages []string     `json:"messages"`
		Results  []model.Item `json:"result"`
	}{
		Status:   "ok",
		Messages: []string{},
		Results:  items,
	}

	js, err := json.Marshal(output)
	if err != nil {
		fmt.Println(err)
		bailOut(w, "error")
		return
	}

	w.Write([]byte(js))
}

func bailOut(w http.ResponseWriter, message string) {

	output := struct {
		Status   string       `json:"status"`
		Messages []string     `json:"messages"`
		Results  []model.Item `json:"result"`
	}{
		Status:   "error",
		Messages: []string{message},
		Results:  []model.Item{},
	}

	js, err := json.Marshal(output)
	if err != nil {
		fmt.Println(err)
		bailOut(w, "error")
		return
	}
	w.Write([]byte(js))
}

func main() {
	// defer the close till after the main function has finished
	defer dbh.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	logger := log.New(os.Stdout, "", 0)

	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":2019"
	}

	s := NewServer(func(s *Server) { s.logger = logger })

	h := &http.Server{Addr: addr, Handler: s}

	go func() {
		logger.Printf("Listening on http://0.0.0.0%s\n", addr)

		if err := h.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()

	sig := <-stop

	logger.Println("\nShutting down the server: ", sig)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	h.Shutdown(ctx)
	logger.Println("Server gracefully stopped")
}
