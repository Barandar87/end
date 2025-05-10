package template

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
	ID              int
	Title           string
	Contents        string
	PublishedOn     string
	URL             string
	CommentID       int
	CommentContents string
}

// Interface declares all desired database operation methods (store and retrieve)
type Interface interface {
	GetNewsTitles() ([]string, error)                           //Extracts and shows the list of news titles (latest to oldest)
	GetNewsItems() ([]NewsItem, error)                          //Extracts news items
	GetNewsItemsByParam(...string) ([]NewsItem, error)          //Extracts news items containing a Param value (string) in their database fields
	AddNews([]NewsItem)                                         //Adds a news item to the database
	AddCommentToNewsItem(CommentItem, NewsItemTitle string)     //Adds a comment to existing news item (parent) to the comment database
	GetCommentsToNewsItem(NewsItemURL string) ([]string, error) //Extracts and shows all the comments to a news item (latest to oldest)
	GetCommentedNews() CommentedNewsItem                        //Extracts and shows the selected news item and all the comments to it
}
