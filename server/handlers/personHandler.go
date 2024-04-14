package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"server/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func AddPerson(c *gin.Context) {
	name := c.PostForm("name")
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	bytes, err := ioutil.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tempFile, err := ioutil.TempFile("", "image.*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, err = tempFile.Write(bytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tempFile.Close()

	_, err = utils.CallPythonCLI("add-person", name, tempFile.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	os.Remove(tempFile.Name())

	htmlResponse := fmt.Sprintf("<h2>%s added successfully</h2>", name)
	c.Data(200, "text/html", []byte(htmlResponse))
}

func SearchPerson(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("Error getting image file: %v", err)
		c.Data(400, "text/html", []byte(fmt.Sprintf("<p>Error: %s</p>", err.Error())))
		return
	}

	tempPath := filepath.Join(os.TempDir(), file.Filename)
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		log.Printf("Error saving image file: %v", err)
		c.Data(500, "text/html", []byte(fmt.Sprintf("<p>Error: %s</p>", err.Error())))
		return
	}

	result, err := utils.CallPythonCLI("search-person", tempPath)
	if err != nil {
		log.Printf("Error calling Python CLI: %v", err)
		c.Data(500, "text/html", []byte(fmt.Sprintf("<p>Error: %s</p>", err.Error())))
		return
	}

	var jsonResult map[string]interface{}
	err = json.Unmarshal([]byte(result), &jsonResult)
	if err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		c.Data(500, "text/html", []byte(fmt.Sprintf("<p>Error: %s</p>", err.Error())))
		return
	}

	mainImageBase64 := strings.TrimPrefix(jsonResult["main_image"].(string), "data:image/jpeg;base64,")
	mainImageData, err := base64.StdEncoding.DecodeString(mainImageBase64)
	if err != nil {
		log.Printf("Error decoding base64 string: %v", err)
		c.Data(500, "text/html", []byte(fmt.Sprintf("<p>Error: %s</p>", err.Error())))
		return
	}

	mainImageDataURL := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(mainImageData)

	htmlResponse := fmt.Sprintf("<h1>%s</h1><img src='%s' alt=''>", jsonResult["Found"], mainImageDataURL)
	c.Data(200, "text/html", []byte(htmlResponse))
}
