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
)

var db *sql.DB
var err error

//Matches columns in table Cars - for the date fields, get a time.Time from string like:
//time1, err := time.Parse("2006-1-2", car1.DatePurchased)
type car struct {
	CarID                int
	LicensePlate         string
	Make                 string
	Model                string
	ModelYear            int
	OdometerReading      float32
	Units                string
	DatePurchased        string
	MileageWhenPurchased float32
	CurrentlyActive      bool
	MileageWhenSold      float64
	DateSold             string
	Nickname             string
}

type fillUp struct {
	PurchaseID       int
	CarID            int
	StationID        int
	PurchaseDate     string
	GallonsPurchased float32
	IsFillUp         bool
	TripMileage      float32
	OdometerReading  float32
	Cost             float32
	Units            string
}

type serviceStation struct {
	StationID int
	Name      string
	Address   string
}

type user struct {
	UserID   int
	Username string
	//TODO: Impliment password hashing so that text passwords are not stored
	Password []byte
}

type repair struct {
	TransactionID   int
	CarID           int
	StationID       int
	PurchaseDate    string
	OdometerReading float32
	Cost            float32
	Description     string
	Units           string
}

var tpl *template.Template

//Appconfig is
type appConfig struct {
	DBConfig   string `json:"dbCon"`
	CertConfig struct {
		Fullchain string `json:"fullchain"`
		PrivKey   string `json:"privkey"`
	} `json:"certconfigs"`
	OAuthConfig struct {
		ClientID     string `json:"clientid"`
		ClientSecret string `json:"clientsecret"`
	} `json:"oauthconfigs"`
}

var config appConfig

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

func main() {

	file, err := ioutil.ReadFile("secret.config.json")
	if err != nil {
		log.Fatalln("config file error")
	}
	json.Unmarshal(file, &config)

	//format found in config file: user:password@tcp(localhost:5555)/dbname?charset=utf8
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
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/about", about)
	//---------Authenticated pages---------------------------------------------
	http.HandleFunc("/viewCars", viewAllCars)
	http.HandleFunc("/viewFillUps", viewFillUps)
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
	//SecretConfigs 1 and 2 are the file path to the fullchain .pem and privkey .pem
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
	io.WriteString(res, "login page")

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
