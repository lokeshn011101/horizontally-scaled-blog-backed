package routes

import (
	"blog/types"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetUser(c *gin.Context) {
	userID := c.Param("user_id")
	pool := c.MustGet("dbPool").(*pgxpool.Pool)
	conn, err := pool.Acquire(c)
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	var user types.Blog_users
	query := fmt.Sprintf(`SELECT id, name, count FROM blog_users WHERE id = %s`, userID)
	err = conn.QueryRow(c, query).Scan(&user.ID, &user.Name, &user.Count)
	conn.Release()
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(200, user)
}

func GetUserBlogs(c *gin.Context) {
	userID := c.Param("user_id")
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
				id =%s`,
		userID)
	rows, _ := conn.Query(c, query)
	rowsData, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Blog_blogs])
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
