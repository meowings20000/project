package mapping

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	dbquery "rethinkdb"

	"gopkg.in/rethinkdb/rethinkdb-go.v6"
)

type Order struct {
	RName         string  `json:"RName" rethinkdb:"RName,omitempty"`
	RAddress      string  `json:"RAddress" rethinkdb:"RAddress,omitempty"`
	RPhone        string  `json:"RPhone" rethinkdb:"RPhone ,omitempty"`
	SName         string  `json:"SName" rethinkdb:"SName ,omitempty"`
	SAddress      string  `json:"SAddress" rethinkdb:"SAddress ,omitempty"`
	SPhone        string  `json:"SPhone" rethinkdb:"SPhone,omitempty"`
	Express       bool    `json:"Express " rethinkdb:"Express ,omitempty"`
	OrderID       string  `json:"OrderID" rethinkdb:"OrderID,omitempty"`
	OrderTime     string  `json:"OrderTime" rethinkdb:"OrderTime"`
	ReceiveTime   string  `json:"ReceiveTime" rethinkdb:"ReceiveTime"`
	DeliveredTime string  `json:"DeliveredTime" rethinkdb:"DeliveredTime"`
	SDriverID     string  `json:"SDriverID" rethinkdb:"SDriverID"`
	RDriverID     string  `json:"RDriverID" rethinkdb:"RDriverID "`
	Weight        float64 `json:"Weight " rethinkdb:"Weight"`
}
type Routes struct {
	V_id   int   `json:"vehicle_id"`
	Routes []int `json:"route"`
}
type Places struct {
	Origin      []Geocodes `json:"origins"`
	Destination []Geocodes `json:"destinations"`
	TravelMode  string     `json:"travelMode"`
}
type Geocodes struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type Respond struct {
	Routess    []Routes `json:"routes"`
	Len        int      `json:"length"`
	Placeslist Places   `json:"places"`
	Orders     []Order  `json:"orders"`
}
type Placerecord struct {
	Placeno  int      `json:"Placeno"`
	Geocodes Geocodes `json:"Geocodes"`
}

func Pythonrun() {
	err := exec.Command("../../map/output/map.exe").Run()
	if err != nil {
		fmt.Println(err)
	}
	print(err)
}

func Details(w http.ResponseWriter, r *http.Request) {
	var exist bool
	var leng int
	file, _ := ioutil.ReadFile("data.json")
	var data []Routes
	_ = json.Unmarshal([]byte(file), &data)
	ffile, _ := ioutil.ReadFile("places.json")
	var places Places
	//var try map[string]interface{}
	_ = json.Unmarshal([]byte(ffile), &places)
	session, _ := dbquery.Connectdb("delivery", "routes")
	for i := 0; i < len(data); i++ {
		query := rethinkdb.DB("delivery").Table("routes").Filter(rethinkdb.Row.Field("V_id").Eq(i)).Distinct().Count()
		result, _ := session.Query(query)
		result.One(&exist)
		if !exist {
			query := rethinkdb.DB("delivery").Table("routes").Insert(data[i])
			session.Query(query)
		} else {
			query := rethinkdb.DB("delivery").Table("routes").Filter(rethinkdb.Row.Field("V_id").Eq(i)).Update(data[i])
			session.Query(query)
		}
	}
	var orders []Order
	query := rethinkdb.DB("delivery").Table("orders")
	result, _ := session.Query(query)
	result.All(&orders)
	session.Closeconn()
	session, _ = dbquery.Connectdb("delivery", "laglogroutes")
	for i := 0; i < len(orders); i++ {
		query := rethinkdb.DB("delivery").Table("laglogroutes").Filter(rethinkdb.Row.Field("Placeno").Eq(i + 1)).Distinct().Count()
		result, _ := session.Query(query)
		result.One(&exist)
		if !exist {
			geo := places.Destination[i]
			p := Placerecord{(i + 1), geo}
			query := rethinkdb.DB("delivery").Table("laglogroutes").Insert(p)
			session.Query(query)

		} else {
			geo := places.Destination[i]
			p := Placerecord{(i + 1), geo}
			query := rethinkdb.DB("delivery").Table("laglogroutes").Filter(rethinkdb.Row.Field("Placeno").Eq(i + 1)).Update(p)
			session.Query(query)

		}
	}
	if !exist {
		geo := places.Origin[0]
		p := Placerecord{0, geo}
		query = rethinkdb.DB("delivery").Table("laglogroutes").Insert(p)
		session.Query(query)
	} else {
		geo := places.Origin[0]
		p := Placerecord{0, geo}
		query := rethinkdb.DB("delivery").Table("laglogroutes").Filter(rethinkdb.Row.Field("Placeno").Eq(0)).Update(p)
		session.Query(query)
	}
	query = rethinkdb.DB("delivery").Table("orders").Count()
	result, _ = session.Query(query)
	result.One(&leng)

	res := Respond{Routess: data, Len: leng, Placeslist: places, Orders: orders}
	w.Header().Set("Content-Type", "application/json")
	repond, _ := json.Marshal(res)
	w.Write(repond)

}
