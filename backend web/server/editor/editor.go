package editor

import (
	"encoding/json"
	"net/http"
	dbquery "rethinkdb"

	"gopkg.in/rethinkdb/rethinkdb-go.v6"
)

type Template struct {
	Id   string      `json:"id"`
	Data interface{} `json:"data"`
}

func Savepage(w http.ResponseWriter, r *http.Request) {
	var consists bool
	session, _ := dbquery.Connectdb("webpages", "webpage")
	var temp Template
	err := json.NewDecoder(r.Body).Decode(&temp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := rethinkdb.DB("webpages").Table("webpage").Filter(rethinkdb.Row.Field("Id").Eq("frontend")).Distinct().Count()
	result, _ := session.Query(query)
	result.One(&consists)
	if consists {
		query := rethinkdb.DB("webpages").Table("webpage").Filter(rethinkdb.Row.Field("Id").Eq("frontend")).Update(temp)
		session.Query(query)
	} else {
		query := rethinkdb.DB("webpages").Table("webpage").Insert(temp)
		session.Query(query)
	}

}

func Loadpage(w http.ResponseWriter, r *http.Request) {
	var consists bool
	session, _ := dbquery.Connectdb("webpages", "webpage")
	var temp Template
	query := rethinkdb.DB("webpages").Table("webpage").Filter(rethinkdb.Row.Field("Id").Eq("frontend")).Distinct().Count()
	result, _ := session.Query(query)
	result.One(&consists)

	if consists {
		query := rethinkdb.DB("webpages").Table("webpage").Filter(rethinkdb.Row.Field("Id").Eq("frontend"))
		result, _ = session.Query(query)
		result.One(&temp)
		jsonBytes, err := json.Marshal(temp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	}

}
