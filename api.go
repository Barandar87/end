package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	postgres "project/pkg/dtbs"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var Filter string

// RequestIDContextKey stores the ID in request context
type RequestIDContextKey struct{}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// establishing the software interfaces for our servers
type NAPI struct {
	NewsDb *postgres.Storage
	router *mux.Router
}
type CAPI struct {
	CommentsDb *postgres.Storage
	router     *mux.Router
}

type GAPI struct {
	NewsDb     *postgres.Storage
	CommentsDb *postgres.Storage
	router     *mux.Router
}

// constructing our API objects
func NewNAPI(NewsDb *postgres.Storage) *NAPI {
	NAPI := NAPI{
		NewsDb: NewsDb,
		router: mux.NewRouter(),
	}
	NAPI.endpoints()
	return &NAPI
}

func NewCAPI(CommentsDb *postgres.Storage) *CAPI {
	CAPI := CAPI{
		CommentsDb: CommentsDb,
		router:     mux.NewRouter(),
	}
	CAPI.endpoints()
	return &CAPI
}

func NewGAPI(NewsDb, CommentsDb *postgres.Storage) *GAPI {
	GAPI := GAPI{
		NewsDb:     NewsDb,
		CommentsDb: CommentsDb,
		router:     mux.NewRouter(),
	}
	GAPI.endpoints()
	return &GAPI
}

// Router functions return our request routers
func (NAPI *NAPI) Router() *mux.Router {
	return NAPI.router
}

func (CAPI *CAPI) Router() *mux.Router {
	return CAPI.router
}

func (GAPI *GAPI) Router() *mux.Router {
	return GAPI.router
}

// registering our general API handlers
func (GAPI *GAPI) endpoints() {
	//getting our commented news
	GAPI.router.HandleFunc("/commentednews/", GAPI.GetCommentedNews).Methods(http.MethodGet)
	// web app
	GAPI.router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webapp"))))
}

// registering our NewsAPI handlers
func (NAPI *NAPI) endpoints() {
	//getting our news
	NAPI.router.HandleFunc("/news/", NAPI.GetNews).Methods(http.MethodGet)
	//getting our news by parameter from path
	NAPI.router.HandleFunc("/news/{f}", NAPI.GetNewsItemsByParam).Methods(http.MethodGet)
	//getting titles for our news
	NAPI.router.HandleFunc("/news/titles/", NAPI.GetNewsTitles).Methods(http.MethodGet)
	// web app
	NAPI.router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webapp"))))
}

// registering our CommentsAPI handlers
func (CAPI *CAPI) endpoints() {
	//getting our news
	CAPI.router.HandleFunc("/comments/{n}", CAPI.GetComments).Methods(http.MethodGet)
	// web app
	CAPI.router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webapp"))))
}

// GetCommentedNews retrieves the desired news item and its comments from our storage collection
func (GAPI *GAPI) GetCommentedNews(w http.ResponseWriter, r *http.Request) {

	lrw := NewLoggingResponseWriter(w)
	//log details
	//Extracting requestID query parameter
	requestID, err := strconv.Atoi(r.URL.Query().Get("request_id"))
	if err != nil || requestID < 1 {
		requestID = rangeIn(int(100000), int(999999999))
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIDContextKey{}, requestID)
		r = r.WithContext(ctx)
		requestTimeStamp := time.Now()
		clientIP := r.RemoteAddr
		respCode := lrw.statusCode
		reqID, ok := r.Context().Value(RequestIDContextKey{}).(string)
		if !ok {
			log.Println("Failed to retrieve Request ID from context.")
		}
		//we now create our destination file with the preestablished name
		//we allow reading and writng
		bReqId := []byte(reqID)
		bReqTimeStamp, err := requestTimeStamp.MarshalBinary()
		if err != nil {
			log.Printf("Encountered an error: %v.", err)
		}
		bClientIP := []byte(clientIP)
		bRespCode := []byte(strconv.Itoa(respCode))
		string := "Here is the log:" + " " + string(bReqId) + " " + string(bReqTimeStamp) + " " + string(bClientIP) + " " + string(bRespCode) + "."
		log.Println(string)
		os.WriteFile("outputFile.txt", []byte(string), os.ModePerm)

		//calling the database Get method and writing it to news variable
		commentedNews := postgres.CommentedNews()
		//transforming the received data to json and sending it to client
		json.NewEncoder(lrw).Encode(commentedNews)
	}
}

// GetNewsTtitles retrieves all the news titles from our storage collection
func (NAPI *NAPI) GetNewsTitles(w http.ResponseWriter, r *http.Request) {

	lrw := NewLoggingResponseWriter(w)
	//log details
	//Extracting requestID query parameter
	requestID, err := strconv.Atoi(r.URL.Query().Get("request_id"))
	if err != nil || requestID < 1 {
		requestID = rangeIn(int(100000), int(999999999))
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIDContextKey{}, requestID)
		r = r.WithContext(ctx)
		requestTimeStamp := time.Now()
		clientIP := r.RemoteAddr
		respCode := lrw.statusCode
		reqID, ok := r.Context().Value(RequestIDContextKey{}).(string)
		if !ok {
			log.Println("Failed to retrieve Request ID from context.")
		}
		//we now create our destination file with the preestablished name
		//we allow reading and writng
		bReqId := []byte(reqID)
		bReqTimeStamp, err := requestTimeStamp.MarshalBinary()
		if err != nil {
			log.Printf("Encountered an error: %v.", err)
		}
		bClientIP := []byte(clientIP)
		bRespCode := []byte(strconv.Itoa(respCode))
		string := "Here is the log:" + " " + string(bReqId) + " " + string(bReqTimeStamp) + " " + string(bClientIP) + " " + string(bRespCode) + "."
		log.Println(string)
		os.WriteFile("outputFile.txt", []byte(string), os.ModePerm)

		//calling the database Get method and writing it to news variable
		titles, _ := NAPI.NewsDb.GetNewsTitles()
		//transforming the received data to json and sending it to client
		json.NewEncoder(lrw).Encode(titles)
	}
}

// GetNews retrieves all the news from our storage collection
func (NAPI *NAPI) GetNews(w http.ResponseWriter, r *http.Request) {
	// Extracting "page" and "limit" query parameters
	Page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || postgres.Page < 1 {
		postgres.Page = 1 // Default to page 1
	} else {
		postgres.Page = Page
	}
	Limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || Limit < 1 {
		postgres.Limit = 10 // Default to 10 items per page
	} else {
		postgres.Limit = Limit
	}
	lrw := NewLoggingResponseWriter(w)
	//log details
	//Extracting requestID query parameter
	requestID, err := strconv.Atoi(r.URL.Query().Get("request_id"))
	if err != nil || requestID < 1 {
		requestID = rangeIn(int(100000), int(999999999))
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIDContextKey{}, requestID)
		r = r.WithContext(ctx)
		requestTimeStamp := time.Now()
		clientIP := r.RemoteAddr
		respCode := lrw.statusCode
		reqID, ok := r.Context().Value(RequestIDContextKey{}).(string)
		if !ok {
			log.Println("Failed to retrieve Request ID from context.")
		}
		//we now create our destination file with the preestablished name
		//we allow reading and writng
		bReqId := []byte(reqID)
		bReqTimeStamp, err := requestTimeStamp.MarshalBinary()
		if err != nil {
			log.Printf("Encountered an error: %v.", err)
		}
		bClientIP := []byte(clientIP)
		bRespCode := []byte(strconv.Itoa(respCode))
		string := "Here is the log:" + " " + string(bReqId) + " " + string(bReqTimeStamp) + " " + string(bClientIP) + " " + string(bRespCode) + "."
		log.Println(string)
		os.WriteFile("outputFile.txt", []byte(string), os.ModePerm)

		//calling the database Get method and writing it to news variable
		news, _ := NAPI.NewsDb.GetNewsItems()
		fmt.Println(news)
		//transforming the received data to json and sending it to client
		json.NewEncoder(lrw).Encode(news)
	}
}

// GetNewsItemsByParam retrieves filtered news from our storage collection
func (NAPI *NAPI) GetNewsItemsByParam(w http.ResponseWriter, r *http.Request) {
	//the filter is grabbed from the "path" part of the request and saved as a variable
	//by using mux.Vars method
	//errors are reported
	//if no errors occur, the http status is returned to the client
	//e.g. "10" from "/news/10"
	Filters := mux.Vars(r)["f"]
	fmt.Println(Filters)
	Filters = "%" + Filters + "%"
	fmt.Println(Filters)

	lrw := NewLoggingResponseWriter(w)
	//log details
	//Extracting requestID query parameter
	requestID, err := strconv.Atoi(r.URL.Query().Get("request_id"))
	if err != nil || requestID < 1 {
		requestID = rangeIn(int(100000), int(999999999))
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIDContextKey{}, requestID)
		r = r.WithContext(ctx)
		requestTimeStamp := time.Now()
		clientIP := r.RemoteAddr
		respCode := lrw.statusCode
		reqID, ok := r.Context().Value(RequestIDContextKey{}).(string)
		if !ok {
			log.Println("Failed to retrieve Request ID from context.")
		}
		//we now create our destination file with the preestablished name
		//we allow reading and writng
		bReqId := []byte(reqID)
		bReqTimeStamp, err := requestTimeStamp.MarshalBinary()
		if err != nil {
			log.Printf("Encountered an error: %v.", err)
		}
		bClientIP := []byte(clientIP)
		bRespCode := []byte(strconv.Itoa(respCode))
		string := "Here is the log:" + " " + string(bReqId) + " " + string(bReqTimeStamp) + " " + string(bClientIP) + " " + string(bRespCode) + "."
		log.Println(string)
		os.WriteFile("outputFile.txt", []byte(string), os.ModePerm)

		//calling the database Get method and writing it to news variable
		filteredNews, _ := NAPI.NewsDb.GetNewsItemsByParam(Filters)
		//transforming the received data to json and sending it to client
		json.NewEncoder(lrw).Encode(filteredNews)
	}
}

// GetCommentsToNewsItem retrieves the specified amount # of the latest news from our storage collection
func (CAPI *CAPI) GetComments(w http.ResponseWriter, r *http.Request) {
	//the ParentID(NewsItem ID) is grabbed from the "path" part of the request and saved as a variable
	//by using mux.Vars method
	//errors are reported
	//if no errors occur, the http status is returned to the client
	//e.g. "10" from "/news/10"
	vars := mux.Vars(r)["n"]
	n, _ := strconv.Atoi(vars)

	lrw := NewLoggingResponseWriter(w)
	//log details
	//Extracting requestID query parameter
	requestID, err := strconv.Atoi(r.URL.Query().Get("request_id"))
	if err != nil || requestID < 1 {
		requestID = rangeIn(int(100000), int(999999999))
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIDContextKey{}, requestID)
		r = r.WithContext(ctx)
		requestTimeStamp := time.Now()
		clientIP := r.RemoteAddr
		respCode := lrw.statusCode
		reqID, ok := r.Context().Value(RequestIDContextKey{}).(string)
		if !ok {
			log.Println("Failed to retrieve Request ID from context.")
		}
		//we now create our destination file with the preestablished name
		//we allow reading and writng
		bReqId := []byte(reqID)
		bReqTimeStamp, err := requestTimeStamp.MarshalBinary()
		if err != nil {
			log.Printf("Encountered an error: %v.", err)
		}
		bClientIP := []byte(clientIP)
		bRespCode := []byte(strconv.Itoa(respCode))
		string := "Here is the log:" + " " + string(bReqId) + " " + string(bReqTimeStamp) + " " + string(bClientIP) + " " + string(bRespCode) + "."
		log.Println(string)
		os.WriteFile("outputFile.txt", []byte(string), os.ModePerm)

		//calling the database Get method and writing it to news variable
		comments, _ := CAPI.CommentsDb.GetCommentsToNewsItem(n)
		//transforming the received data to json and sending it to client
		json.NewEncoder(lrw).Encode(comments)
	}
}

func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
