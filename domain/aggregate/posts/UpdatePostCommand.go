package posts

type UpdatePostCommand struct {
	Uid     int
	Title   string
	Slug    string
	Summary string
	Content string
}
