package main

import (
	"net/http"
	"server/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("ui/*html")
	r.Static("/css", "./ui/css")
	r.Static("/images", "../database/main_images")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"content": "home"})
	})

	r.GET("/add_person", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"content": "add_person"})
	})

	r.GET("/search_person", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"content": "search_person"})
	})

	r.GET("/update_person", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"content": "update_person"})
	})

	r.GET("/search_name", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"content": "search_name"})
	})

	r.POST("/add_person", handlers.AddPerson)

	r.POST("/search_person", handlers.SearchPerson)

	r.POST("/update_person", handlers.UpdatePerson)

	r.POST("/search_name", handlers.SearchName)

	r.Run("192.168.1.6:1907")
}
