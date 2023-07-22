package types

type Blog_users struct {
	ID    int32
	Name  string
	Count int32
}

type Blog_tags struct {
	ID   int32
	Name string
}

type Blog_blogs struct {
	ID         int32
	Title      string
	Text       string
	User_id    int32
	Images     []string
	Is_deleted bool
	Excerpt    string
}

type Blogs_Tags struct {
	Blog_id int32
	Tag_id  int32
}

type Blog_with_tag struct {
	Blog Blog_blogs
	Tag  Blog_tags
}
