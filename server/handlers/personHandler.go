package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"server/utils"

	"github.com/gin-gonic/gin"
)

func handleError(c *gin.Context, err error, statusCode int) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}

func AddPerson(c *gin.Context) {
	name := c.PostForm("name")
	file, err := c.FormFile("image")
	if err != nil {
		handleError(c, err, http.StatusBadRequest)
		return
	}

	src, err := file.Open()
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	defer src.Close()

	tempFile, err := os.CreateTemp("", "image.*")
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name()) // Delete the file after this function returns
	}()

	if _, err := io.Copy(tempFile, src); err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	output, err := utils.CallPythonCLI("add-person", name, tempFile.Name())
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	fmt.Println("Python CLI Input:", name, tempFile.Name())
	fmt.Println("Python CLI output:", output)

	var htmlResponse string
	switch {
	case result["Error"] != "":
		htmlResponse = fmt.Sprintf("<h2 class='error'>Error: %s</h2>", result["Error"])
	case result["Success"] != "":
		htmlResponse = fmt.Sprintf("<h2 class='success'>Success: %s</h2>", result["Success"])
	default:
		htmlResponse = "<h2 class='unknown'>Unknown response from add-person command</h2>"
	}

	c.Data(200, "text/html", []byte(htmlResponse))
}

func SearchPerson(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		handleError(c, err, http.StatusBadRequest)
		return
	}

	tempPath := filepath.Join(os.TempDir(), file.Filename)
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	result, err := utils.CallPythonCLI("search-person", tempPath)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	var jsonResult map[string]interface{}
	if err := json.Unmarshal([]byte(result), &jsonResult); err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	mainImagePath := "/images/" + filepath.Base(jsonResult["main_image"].(string))

	var htmlResponse string
	switch {
	case jsonResult["Error"] != nil && jsonResult["Error"] != "":
		htmlResponse = fmt.Sprintf("<h2 class='error'>Error: %s</h2>", fmt.Sprint(jsonResult["Error"]))
	case jsonResult["Found"] != nil:
		htmlResponse = fmt.Sprintf(`
            <div class='results-grid'>
                <div class='result-card'>
                    <img class='result-image' src='%s' alt=''>
                    <div class='result-name'>
                        <h2>%s</h2>
                    </div>
                </div>
            </div>
        `, mainImagePath, jsonResult["Found"])
	default:
		htmlResponse = "<h2 class='unknown'>Unknown response from search-person command</h2>"
	}

	c.Data(200, "text/html", []byte(htmlResponse))
}

func SearchName(c *gin.Context) {
	name := c.PostForm("name")
	output, err := utils.CallPythonCLI("search-name", name)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	var htmlResponse string
	switch {
	case result["Error"] != nil && result["Error"] != "":
		htmlResponse = fmt.Sprintf("<h2 class='error'>Error: %s</h2>", fmt.Sprint(result["Error"]))
	case result["Success"] != nil:
		successResults := result["Success"].([]interface{})
		htmlResponse += "<div class='results-grid'>"
		for _, successResult := range successResults {
			resultMap := successResult.(map[string]interface{})
			name := resultMap["name"].(string)
			mainImagePath := "/images/" + filepath.Base(resultMap["main_image"].(string))

			htmlResponse += fmt.Sprintf(`
                <div class='result-card'>
                    <img class='result-image' src='%s' alt=''>
                    <div class='result-name'>
                        <h2>%s</h2>
                    </div>
                </div>
            `, mainImagePath, name)
		}
		htmlResponse += "</div>"
	default:
		htmlResponse = "<h2 class='unknown'>Unknown response from search-name command</h2>"
	}

	c.Data(200, "text/html", []byte(htmlResponse))
}

func UpdatePerson(c *gin.Context) {
	name := c.PostForm("name")
	file, err := c.FormFile("image")
	if err != nil {
		handleError(c, err, http.StatusBadRequest)
		return
	}

	src, err := file.Open()
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	defer src.Close()

	tempFile, err := os.CreateTemp("", "image.*")
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	if _, err := io.Copy(tempFile, src); err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	output, err := utils.CallPythonCLI("update-person", name, tempFile.Name())
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	fmt.Println("Python CLI output:", output)

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	var htmlResponse string
	switch {
	case result["Error"] != "":
		htmlResponse = fmt.Sprintf("<h2 class='error'>Error: %s</h2>", result["Error"])
	case result["Success"] != "":
		htmlResponse = fmt.Sprintf("<h2 class='success'>Success: %s</h2>", result["Success"])
	default:
		htmlResponse = "<h2 class='unknown'>Unknown response from update-person command</h2>"
	}

	c.Data(200, "text/html", []byte(htmlResponse))
}
