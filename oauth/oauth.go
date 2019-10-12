package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

//User represents what gets returned by the Google api cobbled with what the program adds.
type User struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	dbID          int
}

type appConfig struct {
	OAuthConfig struct {
		ClientID     string `json:"clientid"`
		ClientSecret string `json:"clientsecret"`
	} `json:"oauthconfigs"`
}

var config appConfig
var oauthConfig *oauth2.Config
var oauthState string
var CurrentUser User

func init() {

	file, err := ioutil.ReadFile("secret.config.json")
	if err != nil {
		log.Fatalln("config file error")
	}
	json.Unmarshal(file, &config)

	fmt.Println("testing!!!!!!:" + config.OAuthConfig.ClientID + "|" + config.OAuthConfig.ClientSecret + "|||")

	oauthConfig = &oauth2.Config{
		ClientID:     config.OAuthConfig.ClientID,
		ClientSecret: config.OAuthConfig.ClientSecret,
		RedirectURL:  "https://cardata.jasonradcliffe.com/success",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

//Login is the handler function that will handle the login with google button href.
func Login(res http.ResponseWriter, req *http.Request) {

	//Each time oauthlogin() is called, a unique, random string gets added to the URL for security
	oauthState = numGenerator()
	url := oauthConfig.AuthCodeURL(oauthState)
	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}

//Success verifies the User info after Google returns from authentication.
func Success(res http.ResponseWriter, req *http.Request) {
	receivedState := req.FormValue("state")

	//Verify that the state parameter is the same coming back from Google as was set when we generated the URL
	if receivedState != oauthState {
		res.WriteHeader(http.StatusForbidden)
	} else {

		//Use the code that Google returns to exchange for an access token
		code := req.FormValue("code")
		token, err := oauthConfig.Exchange(oauth2.NoContext, code)
		check(err)

		//Use the Access token to access the identity API, and get the user info
		response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
		check(err)
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)
		json.Unmarshal(contents, &CurrentUser)

	}

}

func numGenerator() string {
	n := make([]byte, 32)
	rand.Read(n)
	return base64.StdEncoding.EncodeToString(n)
}

func check(err error) {
	if err != nil {
		log.Fatalln("something must have happened: ", err)
	}
}
