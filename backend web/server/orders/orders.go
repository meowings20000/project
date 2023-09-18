package orders

import (
	"encoding/json"
	"fmt"
	"net/http"
	dbquery "rethinkdb"
	"strconv"
	"strings"

	"gopkg.in/rethinkdb/rethinkdb-go.v6"
)

type Order struct {
	RName         string  `json:"RName" rethinkdb:"RName,omitempty"`
	RAddress      string  `json:"RAddress" rethinkdb:"RAddress,omitempty"`
	RPhone        string  `json:"RPhone" rethinkdb:"RPhone ,omitempty"`
	SName         string  `json:"SName" rethinkdb:"SName ,omitempty"`
	SAddress      string  `json:"SAddress" rethinkdb:"SAddress ,omitempty"`
	SPhone        string  `json:"SPhone" rethinkdb:"SPhone,omitempty"`
	Express       bool    `json:"Express" rethinkdb:"Express,omitempty"`
	OrderID       string  `json:"OrderID" rethinkdb:"OrderID,omitempty"`
	OrderTime     string  `json:"OrderTime" rethinkdb:"OrderTime"`
	ReceiveTime   string  `json:"ReceiveTime" rethinkdb:"ReceiveTime"`
	DeliveredTime string  `json:"DeliveredTime" rethinkdb:"DeliveredTime"`
	SDriverID     string  `json:"SDriverID" rethinkdb:"SDriverID"`
	RDriverID     string  `json:"RDriverID " rethinkdb:"RDriverID "`
	Weight        float64 `json:"Weight " rethinkdb:"Weight"`
}

func OrderList(w http.ResponseWriter, r *http.Request) {
	var orders []Order
	session, _ := dbquery.Connectdb("delivery", "orders")
	query := rethinkdb.DB("delivery").Table("orders")
	result, _ := session.Query(query)
	_ = result.All(&orders)
	jsonBytes, _ := json.Marshal(orders)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
	session.Closeconn()
}

func Editorder(w http.ResponseWriter, r *http.Request) {
	var repeated bool

	r.ParseForm()
	session, _ := dbquery.Connectdb("delivery", "orders")
	query := rethinkdb.DB("delivery").Table("orders").Filter(rethinkdb.Row.Field("OrderID").Eq(strings.Join(r.Form["orderid"], " "))).Distinct().Count()
	//query := rethinkdb.DB("delivery").Table("orders").Get("118d2159-316e-444b-bbdb-908fba107ad5").Update(map[string]string{"ReceiveTime": "", "DeliveredTime": ""})
	result, _ := session.Query(query)
	_ = result.One(&repeated)
	if repeated {
		query = rethinkdb.DB("delivery").Table("orders").Filter(
			map[string]string{"OrderID": strings.Join(r.Form["orderid"], " ")}).Update(
			map[string]string{
				"RName": strings.Join(r.Form["RName"], " "), "RAddress": strings.Join(r.Form["RAddress"], " "), "RPhone ": strings.Join(r.Form["RPhone"], " "), "SName ": strings.Join(r.Form["SName"], " "),
				"SAddress ": strings.Join(r.Form["SAdd"], " "), "SPhone": strings.Join(r.Form["SPhone"], " ")})
		_, err := session.Query(query)
		weight, _ := strconv.ParseFloat(strings.Join(r.Form["Weight"], " "), 64)
		query = rethinkdb.DB("delivery").Table("orders").Filter(
			map[string]string{"OrderID": strings.Join(r.Form["orderid"], " ")}).Update(map[string]float64{"Weight": weight})
		session.Query(query)
		fmt.Print(err)
		w.WriteHeader(200)
	} else {
		w.WriteHeader(420)
		w.Write([]byte("Please create an account"))
	}
	session.Closeconn()
}

func Dispatch(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var driver []int
	var routeid []int
	var orders []Order
	var route []int
	routestr := strings.Join(r.Form["routeid"], " ")
	s := strings.Split(routestr, ",")
	for _, element := range s {
		tmpid, _ := strconv.Atoi(element)
		routeid = append(routeid, tmpid)
	}
	for i := 0; i < len(routeid); i++ {
		str := "DriverID" + strconv.Itoa(i)
		tmpdriver, _ := strconv.Atoi(strings.Join(r.Form[str], " "))
		driver = append(driver, tmpdriver)
	}
	session, _ := dbquery.Connectdb("delivery", "orders")
	query := rethinkdb.DB("delivery").Table("orders")
	result, _ := session.Query(query)
	_ = result.All(&orders)
	session.Closeconn()
	session, _ = dbquery.Connectdb("delivery", "routes")
	query = rethinkdb.DB("delivery").Table("routes").Update(map[string]int{"DriverID": -1})
	result, _ = session.Query(query)
	x := 0
	for i := range routeid {
		session, _ = dbquery.Connectdb("delivery", "routes")
		query := rethinkdb.DB("delivery").Table("routes").Filter(map[string]int{"V_id": routeid[i]}).Field("Routes")
		result, _ := session.Query(query)
		_ = result.One(&route)
		for _, element := range route {
			if element != 0 {
				query = rethinkdb.DB("delivery").Table("orders").Filter(map[string]string{"OrderID": orders[element-1].OrderID}).Update(map[string]int{"RDriverID ": driver[x]})
				session.Query(query)

			}
		}
		query = rethinkdb.DB("delivery").Table("routes").Filter(map[string]int{"V_id": routeid[i]}).Update(map[string]int{"DriverID": driver[x]})
		session.Query(query)
		x++

	}
	w.WriteHeader(200)
}
