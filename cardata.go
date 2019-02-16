package main

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
