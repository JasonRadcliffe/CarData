package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jasonradcliffe/cardata/oauth"
	"golang.org/x/oauth2"
)

var db *sql.DB
var err error

type appConfig struct {
	DBConfig   string `json:"dbCon"`
	CertConfig struct {
		Fullchain string `json:"fullchain"`
		PrivKey   string `json:"privkey"`
	} `json:"certconfigs"`
}

var tpl *template.Template
var config appConfig
var oauthconfig *oauth2.Config
var oauthstate string
var currentUser oauth.User

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))

	//Open the secret configuration file, and load the info as specified by appConfig struct
	file, err := ioutil.ReadFile("secret.config.json")
	if err != nil {
		log.Fatalln("config file error")
	}
	json.Unmarshal(file, &config)
}

func main() {

	//format needed by sql driver: [username]:[password]@tcp([address:port])/[dbname]?charset=utf8
	db, err = sql.Open("mysql", config.DBConfig)
	check(err)
	defer db.Close()

	//Check the connection to the database - If the credentials are wrong this will err out
	err = db.Ping()
	check(err)

	car1 := getCarFromID(7)
	fmt.Println(car1)

	//Routes-------------------------------------------------------------------
	//---------Unauthenticated pages-------------------------------------------
	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
	http.HandleFunc("/oauthlogin", oauth.Login)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/about", about)
	//---------Authenticated pages---------------------------------------------
	http.HandleFunc("/viewCars", viewAllCars)
	http.HandleFunc("/viewFillUps", viewFillUps)
	http.HandleFunc("/success", oauth.Success)
	//--------------------------------------------------------End Routes-------

	//Server Setup and Config--------------------------------------------------
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	srv := &http.Server{
		Addr:         ":443",
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	//From the config file: the file path to the fullchain .pem and privkey .pem
	log.Fatalln(srv.ListenAndServeTLS(config.CertConfig.Fullchain, config.CertConfig.PrivKey))
	//-----------------------------------------------End Server Setup and Config---
}

func check(err error) {
	if err != nil {
		log.Fatalln("something must have happened: ", err)
	}
}

//Route Functions-----------------------------------------------------------
//--------------Unauthenticated Pages---------------------------------------
func index(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, "homepage")
}

func login(res http.ResponseWriter, req *http.Request) {
	//The Login page for the app - contains a "Login with Google" button
	io.WriteString(res, `<a href="/oauthlogin"> Login with Google </a>`)
}

func signup(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, "signup page")
	tpl.ExecuteTemplate(res, "signup.gphtml", nil)

}
func about(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, "about page")

}

//-----------Authenticated Pages------------------------------------------

func viewAllCars(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, "current user email:"+oauth.CurrentUser.Email)
	rows, err := db.Query(`SELECT * FROM Car2;`)
	check(err)
	defer rows.Close()

	s := "RETRIEVED RECORDS:\n"

	for rows.Next() {
		var car1 car
		var sMileageWhenSold sql.NullFloat64
		var sDateSold, sNickname sql.NullString

		err = rows.Scan(&car1.CarID, &car1.LicensePlate, &car1.Make, &car1.Model, &car1.ModelYear,
			&car1.OdometerReading, &car1.Units, &car1.DatePurchased, &car1.MileageWhenPurchased,
			&car1.CurrentlyActive, &sMileageWhenSold, &sDateSold, &sNickname)
		check(err)
		//s += "Car1:" + car1

		if sMileageWhenSold.Valid {
			car1.MileageWhenSold = sMileageWhenSold.Float64
		}
		if sDateSold.Valid {
			car1.DateSold = sDateSold.String
		}
		if sNickname.Valid {
			car1.Nickname = sNickname.String
		}

		time1, err := time.Parse("2006-1-2", car1.DatePurchased)
		check(err)
		fmt.Println("Car was purchased in the month of:" + strconv.Itoa(int(time1.Month())))

		s += fmt.Sprint(car1) + "\n"
	}
	fmt.Fprintln(res, s)
}

func viewFillUps(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, "{insert viewCars code here!}")
	rows, err := db.Query(`SELECT LicensePlate FROM Car2;`)
	check(err)
	defer rows.Close()

	// data to be used in query
	var s, license string
	s = "RETRIEVED RECORDS:\n"

	// query
	for rows.Next() {
		err = rows.Scan(&license)
		check(err)
		s += license + "\n"
	}
	fmt.Fprintln(res, s)
}

//---------------------------------------------End Route Functions------

func getCarFromID(carID int) car {
	rows, err := db.Query("SELECT * FROM Car2 WHERE CarID =" + strconv.Itoa(carID) + ";")
	check(err)
	defer rows.Close()

	rows.Next()
	var car1 car
	var sMileageWhenSold sql.NullFloat64
	var sDateSold, sNickname sql.NullString

	err = rows.Scan(&car1.CarID, &car1.LicensePlate, &car1.Make, &car1.Model, &car1.ModelYear,
		&car1.OdometerReading, &car1.Units, &car1.DatePurchased, &car1.MileageWhenPurchased,
		&car1.CurrentlyActive, &sMileageWhenSold, &sDateSold, &sNickname)
	check(err)

	if sMileageWhenSold.Valid {
		car1.MileageWhenSold = sMileageWhenSold.Float64
	}
	if sDateSold.Valid {
		car1.DateSold = sDateSold.String
	}
	if sNickname.Valid {
		car1.Nickname = sNickname.String
	}

	return car1
}
