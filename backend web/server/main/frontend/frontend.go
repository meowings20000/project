package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	dbquery "rethinkdb"
	"strings"
	"time"

	"gopkg.in/rethinkdb/rethinkdb-go.v6"
)

func RandomString(n int) string {

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	xrand := int64(rand.Intn(254476375425))
	rand.Seed(time.Now().Unix() + xrand)
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}

	return string(s)

}

type respond struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
}

type HandlerFunc func(http.ResponseWriter, *http.Request)
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

func main() {

	order := func(w http.ResponseWriter, r *http.Request) {
		var order Order
		r.ParseForm()
		order.RName = strings.Join(r.Form["RName"], "")
		order.RAddress = strings.Join(r.Form["RAdd"], "")
		order.RPhone = strings.Join(r.Form["RPhone"], "")
		order.SName = strings.Join(r.Form["SName"], "")
		order.SAddress = strings.Join(r.Form["SAdd"], "")
		order.SPhone = strings.Join(r.Form["SPhone"], "")
		if strings.Join(r.Form["Express"], " ") == "True" {
			order.Express = true
		} else {
			order.Express = false
		}
		order.OrderID = RandomString(16)
		now := time.Now()
		order.OrderTime = now.Format("2006-01-02 15:04:05")
		session, _ := dbquery.Connectdb("delivery", "orders")
		session.Insert(order)

	}
	qrcode := func(w http.ResponseWriter, r *http.Request) {
		var order bool

		r.ParseForm()
		session, _ := dbquery.Connectdb("delivery", "orders")
		check := strings.Join(r.Form["orderid"], " ")
		query := rethinkdb.DB("delivery").Table("orders").Filter(map[string]string{"OrderID": check}).Count().Eq(1)
		result, _ := session.Query(query)
		_ = result.One(&order)
		if order {
			w.Header().Set("Content-Type", "application/json")
			repond := respond{Status: "success", Message: check}
			jsonBytes, _ := json.Marshal(repond)
			w.Write(jsonBytes)
		} else {
			w.Header().Set("Content-Type", "application/json")
			repond := respond{Status: "fail", Message: "Wrong ID"}
			jsonBytes, _ := json.Marshal(repond)
			w.Write(jsonBytes)
		}
		session.Closeconn()
	}
	orderhandler := http.HandlerFunc(order)
	qrhandler := http.HandlerFunc(qrcode)
	http.Handle("/", http.FileServer(http.Dir("../../../frontend")))
	http.Handle("/order", orderhandler)
	http.Handle("/qrcode", qrhandler)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
