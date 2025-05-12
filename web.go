package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"project/pkg/api"
	postgres "project/pkg/dtbs"
	"project/pkg/rss"
	"time"

	"github.com/gorilla/mux"
)

type JSONconfig struct {
	URLs           []string `json:"rss"`
	TraversePeriod int      `json:"request_period"`
}

var Comment1 postgres.Comment
var Comment2 postgres.Comment
var Comment3 postgres.Comment

func main() {
	//initialising the database
	newsDatabase := postgres.ConnectNews()
	commentDatabase := postgres.ConnectComments()

	//initialising the APIs for our databases
	generalApi := api.NewGAPI(newsDatabase, commentDatabase)
	newsApi := api.NewNAPI(newsDatabase)
	commentsApi := api.NewCAPI(commentDatabase)

	//creating comments and sending them to the database
	Comment1 = postgres.Comment{ParentID: 13, Contents: ")) Все по плану февраля 2022.......(((( Пора водкой платить а не руБЛЕВЫМИ лимонами (( фантиками..А если не платить???? А как Чеченские ДВЕ за сигареты и паек ((( Слабо..", PublishedOn: "2025-05-03", URL: "https://lenta.ru/comments/news/2025/05/03/vs-rossii-udarili-po-sobravshimsya-na-proryv-v-bryanskoy-oblasti-silam-vsu/", Allowed: true}
	Comment2 = postgres.Comment{ParentID: 13, Contents: "Курская, Белгородская,Брянская... мыколы, вам не надоело как горох об стену? Сначала свои территории освободите, потом за чужие беритесь.", PublishedOn: "2025-05-03", URL: "https://lenta.ru/comments/news/2025/05/03/vs-rossii-udarili-po-sobravshimsya-na-proryv-v-bryanskoy-oblasti-silam-vsu/", Allowed: true}
	Comment3 = postgres.Comment{ParentID: 13, Contents: "Удивительно, но два документа по сдаче недр штатам засекретили не только он народу Украины, который облапошивают, но и от верховной Рады! Цирк с конями!", PublishedOn: "2025-05-03", URL: "https://lenta.ru/comments/news/2025/05/03/vs-rossii-udarili-po-sobravshimsya-na-proryv-v-bryanskoy-oblasti-silam-vsu/", Allowed: true}

	r := mux.NewRouter()
	r.HandleFunc("/api/comments", commentHandler).Methods("POST")

	var commentsChecked []postgres.Comment
	commentsChecked = append(commentsChecked, Comment1, Comment2, Comment3)

	postgres.ConnectComments().AddComment(commentsChecked)

	//opening our configuration file from the local folder
	file, err := os.ReadFile("./config.json")
	if err != nil {
		log.Fatalf("Could not read the file because of %v./n", err)
	}
	// we read the configuration file and store its contents as a struct
	var Config JSONconfig
	err = json.Unmarshal([]byte(file), &Config)
	if err != nil {
		log.Fatalf("Could not decode the contents because of %v./n", err)
	}
	// log.Println(config)

	newsCh := make(chan []postgres.NewsItem)
	newsErrCh := make(chan error)

	//we call the GET method of HTTP one by one on each url from our configuration file
	//responses are read and the contents are placed into fields of news items
	for _, url := range Config.URLs {
		go parseWebUrl(url, newsCh, newsErrCh, Config.TraversePeriod)
	}

	//this channel sends news to their database
	go func() {
		for news := range newsCh {
			newsDatabase.AddNews(news)
		}
	}()
	//this channel displays the errors from reading news
	go func() {
		for err := range newsErrCh {
			log.Printf("Got an error %v while reading news./n", err)
		}
	}()

	//launching the network service and the HTTP server on local IPs on port 80
	//requests are handed over to the router for processing
	err = http.ListenAndServe(":80", newsApi.Router())
	if err != nil {
		log.Fatalf("Network service failed because of: %v", err)
	}

	err = http.ListenAndServe(":80", commentsApi.Router())
	if err != nil {
		log.Fatalf("Network service failed because of: %v", err)
	}

	err = http.ListenAndServe(":80", generalApi.Router())
	if err != nil {
		log.Fatalf("Network service failed because of: %v", err)
	}
}

// Making our function to send to channel decoded news and errors
// in an endless for loop
func parseWebUrl(url string, newsCh chan<- []postgres.NewsItem, errCh chan<- error, period int) {
	for {
		news, err := rss.ParseURL(url)
		if err != nil {
			errCh <- err
		}
		newsCh <- news
		time.Sleep(time.Minute * time.Duration(period))
	}
}

func commentHandler(w http.ResponseWriter, r *http.Request) {
	//reading the request
	var comment postgres.Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil || comment.Contents == "" {
		http.Error(w, "Invalid comment body", http.StatusBadRequest)
		return
	}
	//re-routing the request to censor
	censorshipURL := "http://localhost/check"
	censorshipPayload, _ := json.Marshal(map[string]string{"text": comment.Contents})
	resp, err := http.Post(censorshipURL, "application/json", bytes.NewBuffer(censorshipPayload))
	if err != nil {
		log.Println("Error contacting censorship service:", err)
		http.Error(w, "Failed to validate comment", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		http.Error(w, "Comment rejected by censorship service", http.StatusBadRequest)
		return
	} else if resp.StatusCode != http.StatusOK {
		http.Error(w, "Censorship service error", http.StatusInternalServerError)
		return
	}
}
