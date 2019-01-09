package main

import (
	"database/sql"
	"fmt"
	"io"
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
	MileageWhenSold      float32
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
	Password string
}

func main() {

	//user:password@tcp(localhost:5555)/dbname?charset=utf8
	db, err = sql.Open("mysql", "")
	check(err)
	defer db.Close()

	err = db.Ping()
	check(err)

	http.HandleFunc("/viewCars", viewAllCars)
	http.HandleFunc("/viewFillUps", viewFillUps)
	http.ListenAndServe(":8080", nil)
}

func check(err error) {
	if err != nil {
		log.Fatalln("something must have happened: ", err)
	}
}

func viewAllCars(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, "{insert viewCars code here!}")
	rows, err := db.Query(`SELECT CarID, DatePurchased FROM Car2;`)
	check(err)
	defer rows.Close()

	// data to be used in query
	var car1 car
	//var carID int
	//var s, license string
	s := "RETRIEVED RECORDS:\n"

	// query
	for rows.Next() {
		err = rows.Scan(&car1.CarID, &car1.DatePurchased)
		check(err)
		//s += "Car1:" + car1
		time1, err := time.Parse("2006-1-2", car1.DatePurchased)
		check(err)
		fmt.Println("Car was purchased in the month of:" + strconv.Itoa(int(time1.Month())))

		s += fmt.Sprint(car1) + "\n"
	}
	fmt.Fprintln(res, s)
}

func getCarFromID(carID int) {

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
