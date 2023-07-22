package routes

import (
	"blog/types"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetBlog(c *gin.Context) {
	blogID := c.Param("blog_id")
	pool := c.MustGet("dbPool").(*pgxpool.Pool)
	conn, err := pool.Acquire(c)
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	var blog types.Blog_blogs
	query := fmt.Sprintf(`
		SELECT
				id,
				title,
				text,
				images,
				user_id,
				is_deleted,
				excerpt
		FROM
				blog_blogs
		WHERE 
				id = %s`, blogID)
	err = conn.QueryRow(c, query).Scan(&blog.ID, &blog.Title, &blog.Text, &blog.Images, &blog.User_id, &blog.Is_deleted, &blog.Excerpt)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(200, blog)
	conn.Release()
}

func DeleteBlog(c *gin.Context) {
	blogID := c.Param("blog_id")
	pool := c.MustGet("dbPool").(*pgxpool.Pool)
	conn, err := pool.Acquire(c)
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	tx, err := conn.Begin(c)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	defer tx.Rollback(c)

	_, err = tx.Exec(c, fmt.Sprintf(`
		UPDATE
				BLOG_BLOGS
		SET
				IS_DELETED = TRUE
		WHERE
				ID = %s`, blogID))
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	if err := tx.Commit(c); err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.Status(200)
	_, err = conn.Exec(c, fmt.Sprintf(`
		DELETE FROM
			blog_blogs
		WHERE
			id = %s`, blogID))
	if err != nil {
		fmt.Println(err)
	}
	conn.Release()
}

func UpdateBlogTags(c *gin.Context) {
	blogID := c.Param("blog_id")
	pool := c.MustGet("dbPool").(*pgxpool.Pool)
	conn, err := pool.Acquire(c)
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	oldTags := c.DefaultQuery("oldTags", "")
	newTagsQuery := c.DefaultQuery("newTags", "")
	newTags := strings.Split(newTagsQuery, ",")

	tx, err := conn.Begin(c)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	defer tx.Rollback(c)

	_, err = tx.Exec(c, fmt.Sprintf(`
		DELETE FROM
			blogs_tags
		WHERE
			tag_id IN (%s)`, oldTags))
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	var values string
	for _, tag := range newTags {
		values += fmt.Sprintf(`(%s, %s),`, blogID, tag)
	}
	values = strings.TrimSuffix(values, ",")
	_, err = tx.Exec(c, fmt.Sprintf(`
		INSERT INTO
			blogs_tags (blog_id, tag_id)
		VALUES
			%s`, values))
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	if err := tx.Commit(c); err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.Status(200)
	conn.Release()
}

func CreateBlog(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("user_id"))
	pool := c.MustGet("dbPool").(*pgxpool.Pool)
	conn, err := pool.Acquire(c)
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	var blog types.Blog_blogs
	if err := c.BindJSON(&blog); err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
	}

	tx, err := conn.Begin(c)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	defer tx.Rollback(c)

	_, err = tx.Exec(c, fmt.Sprintf(`
		INSERT INTO blog_blogs (title, text, user_id, excerpt)
				VALUES ('%s', '%s', %d, '%s')`,
		blog.Title, blog.Text, userID, blog.Excerpt))
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	_, err = tx.Exec(c, fmt.Sprintf(`
		UPDATE
				blog_users
		SET
				count = count + 1
		WHERE
				ID = %d`, userID))
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	if err := tx.Commit(c); err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	conn.Release()
}

func GetUserFeed(c *gin.Context) {
	userID := c.Query("userId")
	pool := c.MustGet("dbPool").(*pgxpool.Pool)
	conn, err := pool.Acquire(c)
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	query := fmt.Sprintf(`
		SELECT
				t.id as tag_id,
				t.name as tag_name,
				b.id,
				b.title,
				b.text,
				b.images,
				b.user_id,
				b.is_deleted,
				b.excerpt
		FROM
				blog_blogs b
				JOIN blogs_tags bt ON b.id = bt.blog_id
				JOIN blog_tags t ON t.id = bt.tag_id
				JOIN (
						SELECT
								id,
								interested_tags
						FROM
								blog_users
						WHERE
								id = %s
				) ut ON t.id = ANY(ut.interested_tags)
		WHERE
				b.user_id != %s
		LIMTI 25`, userID, userID)
	rows, _ := conn.Query(c, query)
	rowsData, err := rowsToBlogs(rows)
	conn.Release()
	if err != nil {
		fmt.Printf("CollectRows error: %v", err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(200, rowsData)
}

func rowsToBlogs(rows pgx.Rows) ([]types.Blog_with_tag, error) {
	var blogs []types.Blog_with_tag
	for rows.Next() {
		var blog types.Blog_with_tag
		err := rows.Scan(&blog.Tag.ID, &blog.Tag.Name, &blog.Blog.ID, &blog.Blog.Title, &blog.Blog.Text, &blog.Blog.Images, &blog.Blog.User_id, &blog.Blog.Is_deleted, &blog.Blog.Excerpt)
		if err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return blogs, nil
}
