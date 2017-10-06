package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
  "fmt"
  "net/url"
	"io/ioutil"
	"encoding/json"
	"os"
)

var DB = make(map[string]string)

type Response struct {
    access_token string
		client_id string
		expires_in string
		refresh_token string
		token_type string
		user_id string
}

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/images", "./images")
	r.Static("/css", "./css")

	r.GET("/", func(c *gin.Context) {
		link := `https://auth.getmondo.co.uk/?`
		link += `client_id=` + os.Getenv("monzo_client_id")
		link += `&redirect_uri=http://localhost:8080/statements&`
		link += `response_type=code&state=12345`
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Statements",
			"button": "Connect Account",
			"link": link,
		})
	})

	r.GET("/statements", func(c *gin.Context) {
		form := url.Values{
			"grant_type": {"authorization_code"},
			"client_id": {os.Getenv("monzo_client_id")},
			"client_secret": {os.Getenv("monzo_client_secret")},
			"redirect_uri": {"http://localhost:8080/statements"},
			"code": {c.Query("code")},
		}

		resp, _ := http.PostForm("https://api.monzo.com/oauth2/token", form)
		var access_token string = ""
		defer resp.Body.Close()


		if resp.StatusCode == 200 {
		    bodyBytes, _ := ioutil.ReadAll(resp.Body)
		    bodyString := string(bodyBytes)
 				var data map[string]interface{}
 				err := json.Unmarshal([]byte(bodyString), &data)
 				if err != nil {
		 			panic(err)
 				}
 				access_token = data["access_token"].(string)
		}

		req, _ := http.NewRequest("GET", "https://api.monzo.com/accounts", nil)
		q := req.URL.Query()
		req.Header.Set("Authorization", "Bearer " + access_token)
		req.URL.RawQuery = q.Encode()
		client := &http.Client{}
		response, _ := client.Do(req)
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(bodyBytes))

		c.HTML(http.StatusOK, "authorised.tmpl", gin.H{
			"accounts": "Good!",
		})
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := DB[user]
		if ok {
			c.JSON(200, gin.H{"user": user, "value": value})
		} else {
			c.JSON(200, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			DB[user] = json.Value
			c.JSON(200, gin.H{"status": "ok"})
		}
	})

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
