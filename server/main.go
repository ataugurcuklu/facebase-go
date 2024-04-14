package main

import (
	"net/http"
	"server/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("ui/*html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"content": "home"})
	})

	r.GET("/add_person", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"content": "add_person"})
	})

	r.GET("/search_person", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"content": "search_person"})
	})

	r.POST("/add_person", handlers.AddPerson)

	r.POST("/search_person", handlers.SearchPerson)

	r.Run()
}
