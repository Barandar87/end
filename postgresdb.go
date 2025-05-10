package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

type NewsItem struct {
	ID          int
	Title       string
	Contents    string
	PublishedOn string
	URL         string
}

type Comment struct {
	ID          int
	ParentID    int //news item ID
	Contents    string
	PublishedOn string
	URL         string
	Allowed     bool
}

type CommentedNewsItem struct {
	ID          int
	Title       string
	Contents    string
	PublishedOn string
	URL         string
	Comments    []Comment
}

var Page int
var Limit int

// NewsConnectionString func creates the string holding parameters
// to establish the database connection
// parameters are thus applied without revealing them to the end user
func NewsConnectionString() string {
	os.Setenv("CONN_STR_NEWS", "postgres://postgres:zwh15lhI@localhost:5432/postgres?sslmode=disable")
	connStr, status := os.LookupEnv("CONN_STR_NEWS")
	if !status {
		log.Fatalln("Missing environment variable CONN_STR_NEWS.")
	}
	return connStr
}

// CommentConnectionString func creates the string holding parameters
// to establish the database connection
// parameters are thus applied without revealing them to the end user
func CommentConnectionString() string {
	os.Setenv("CONN_STR_COMMENTS", "postgres://postgres:zwh15lhI@localhost:5432/comments?sslmode=disable")
	connStr, status := os.LookupEnv("CONN_STR_COMMENTS")
	if !status {
		log.Fatalln("Missing environment variable CONN_STR_COMMENTS.")
	}
	return connStr
}

// ConnectNews func creates a connection pool and reports the error logs (if present) from each step.
// If all the steps worked correctly, a pointer to the database is returned.
func ConnectNews() *Storage {
	connStr := NewsConnectionString()
	dbNewspool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to news database because of %s./n", err)
	}
	if err = dbNewspool.Ping(context.Background()); err != nil {
		log.Fatalf("Cannot ping news database because of %s./n", err)
	}
	NS := Storage{
		db: dbNewspool,
	}
	// log.Println("Successfully connected to news database and pinged it.")
	return &NS
}

// ConnectComments func creates a connection pool and reports the error logs (if present) from each step.
// If all the steps worked correctly, a pointer to the database is returned.
func ConnectComments() *Storage {
	connStr := CommentConnectionString()
	dbCommentspool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to comments database because of %s./n", err)
	}
	if err = dbCommentspool.Ping(context.Background()); err != nil {
		log.Fatalf("Cannot ping comments database because of %s./n", err)
	}
	CS := Storage{
		db: dbCommentspool,
	}
	// log.Println("Successfully connected to comments database and pinged it.")
	return &CS
}

// AddNews func stores news items and reports the error logs (if present) from each step.
// If all the steps worked correctly, a nil error is returned.
func (NS *Storage) AddNews(news []NewsItem) error {
	for _, item := range news {
		_, err := NS.db.Exec(context.Background(), `INSERT INTO "news" ("title", "contents", "publishing_date", "url") values ($1, $2, $3, $4)`, item.Title, item.Contents, item.PublishedOn, item.URL)
		if err != nil {
			log.Fatalf("Failed to update rows because of %s./n", err)
		}
		log.Println("Successfully added a news item.")
	}
	return nil
}

func (NS *Storage) GetNewsItemsByParam(filter string) ([]NewsItem, error) {
	var FilteredNewsItems []NewsItem
	filter = "%" + filter + "%"
	rows, err := NS.db.Query(context.Background(), `SELECT * FROM news WHERE title ILIKE ($1);`, filter)
	if err != nil {
		log.Fatalf("Database query failed because of %s./n", err)
	}
	for rows.Next() {
		var n NewsItem
		err = rows.Scan(&n.ID, &n.Title, &n.Contents, &n.PublishedOn, &n.URL)
		if err != nil {
			log.Fatalf("Failed to retrieve rows because of %s./n", err)
		}
		temp := NewsItem{n.ID, n.Title, n.Contents, n.PublishedOn, n.URL}
		FilteredNewsItems = append(FilteredNewsItems, temp)
		if err := rows.Err(); err != nil {
			log.Fatalf("The following error encountered while iterating over rows: %s./n", err)
		}
	}
	if len(FilteredNewsItems) == 0 {
		fmt.Println("Nothing received!")
	}
	defer rows.Close()
	fmt.Println(FilteredNewsItems)
	return FilteredNewsItems, nil
}

func (NS *Storage) GetNewsTitles() ([]string, error) {
	var titles []string
	rows, err := NS.db.Query(context.Background(), `SELECT id, title
	FROM news
	ORDER BY id DESC`)
	if err != nil {
		log.Fatalf("Database query failed because of %s./n", err)
	}
	for rows.Next() {
		var n NewsItem
		err = rows.Scan(&n.ID, &n.Title)
		if err != nil {
			log.Fatalf("Failed to retrieve rows because of %s./n", err)
		}
		temp := n.Title
		titles = append(titles, temp)
		if err := rows.Err(); err != nil {
			log.Fatalf("The following error encountered while iterating over rows: %s./n", err)
		}
	}
	defer rows.Close()
	return titles, nil
}

func (NS *Storage) GetNewsItems() ([]NewsItem, error) {

	// Calculate the OFFSET
	offset := (Page - 1) * Limit
	var news []NewsItem
	rows, err := NS.db.Query(context.Background(), `SELECT id, title, contents, publishing_date, URL
	FROM news
	ORDER BY id DESC
	LIMIT $1 OFFSET $2`, Limit, offset)
	if err != nil {
		log.Fatalf("Database query failed because of %s./n", err)
	}
	for rows.Next() {
		var n NewsItem
		err = rows.Scan(&n.ID, &n.Title, &n.Contents, &n.PublishedOn, &n.URL)
		if err != nil {
			log.Fatalf("Failed to retrieve rows because of %s./n", err)
		}
		temp := NewsItem{n.ID, n.Title, n.Contents, n.PublishedOn, n.URL}
		news = append(news, temp)
		if err := rows.Err(); err != nil {
			log.Fatalf("The following error encountered while iterating over rows: %s./n", err)
		}
	}
	defer rows.Close()
	return news, nil
}

// AddComment func stores news items and reports the error logs (if present) from each step.
// If all the steps worked correctly, a nil error is returned.
func (CS *Storage) AddComment(comment []Comment) error {
	for _, item := range comment {
		_, err := CS.db.Exec(context.Background(), `INSERT INTO "comments" ("parent_id", "creation_date", "contents", "url") values ($1, $2, $3, $4)`, item.ParentID, item.PublishedOn, item.Contents, item.URL)
		if err != nil {
			log.Fatalf("Failed to update rows because of %s./n", err)
		}
		log.Printf("Successfully added a comment to News item %d.", item.ParentID)
	}
	return nil
}

func (CS *Storage) GetCommentsToNewsItem(NewsItemID int) ([]Comment, error) {
	var comments []Comment
	rows, err := CS.db.Query(context.Background(), `SELECT * FROM comments WHERE parent_id = $1;`, NewsItemID)
	if err != nil {
		log.Fatalf("Database query failed because of %s./n", err)
	}
	for rows.Next() {
		var c Comment
		err = rows.Scan(&c.ID, &c.ParentID, &c.Contents, &c.PublishedOn, &c.URL)
		if err != nil {
			log.Fatalf("Failed to retrieve rows because of %s./n", err)
		}
		temp := Comment{ID: c.ID, ParentID: c.ParentID, Contents: c.Contents, PublishedOn: c.PublishedOn, URL: c.URL}
		comments = append(comments, temp)
		if err := rows.Err(); err != nil {
			log.Fatalf("The following error encountered while iterating over rows: %s./n", err)
		}
	}
	defer rows.Close()
	return comments, nil
}

func CommentedNews() CommentedNewsItem {
	time.Sleep(5 * time.Second)
	NewsItems, _ := ConnectNews().GetNewsItemsByParam("ВС России поразили собравшиеся на прорыв в Брянской области силы ВСУ")
	var NewsItem NewsItem
	for _, item := range NewsItems {
		NewsItem = item
	}

	CommentItems, _ := ConnectComments().GetCommentsToNewsItem(13)
	var CommentItem Comment
	var Comments []Comment
	for _, item := range CommentItems {
		CommentItem = item
		temp := Comment{ID: CommentItem.ID, ParentID: CommentItem.ParentID, Contents: CommentItem.Contents}
		Comments = append(Comments, temp)
	}
	var NewsItemWithComments CommentedNewsItem
	NewsItemWithComments.ID = NewsItem.ID
	NewsItemWithComments.Title = NewsItem.Title
	NewsItemWithComments.Contents = NewsItem.Contents
	NewsItemWithComments.PublishedOn = NewsItem.PublishedOn
	NewsItemWithComments.URL = NewsItem.URL
	NewsItemWithComments.Comments = Comments
	return NewsItemWithComments
}
