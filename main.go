package main

import (
	"blog/routes"
	"blog/sql"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func DBMiddleware(dbPool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("dbPool", dbPool)
		ctx.Next()
	}
}

func main() {
	fmt.Println(os.Args)
	// SetupFakeData()

	router := gin.Default()
	router.Use(DBMiddleware(sql.ConnectToPostgres()))

	v1 := router.Group("/v1")
	{
		users := v1.Group("/users")
		{
			users.GET("/:user_id", routes.GetUser)
			users.GET("/:user_id/blogs", routes.GetUserBlogs)
		}
		blogs := v1.Group("/blogs")
		{
			blogs.GET("/", routes.GetUserFeed)
			blogs.GET("/:blog_id", routes.GetBlog)
			blogs.DELETE("/:blog_id", routes.DeleteBlog)
			blogs.PUT("/:blog_id/:user_id/tags", routes.UpdateBlogTags)
			blogs.POST("/:user_id/new", routes.CreateBlog)
		}
	}

	router.Run(fmt.Sprintf("localhost:%s", os.Args[1]))
}
