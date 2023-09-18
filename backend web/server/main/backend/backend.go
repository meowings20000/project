package main

import (
	accounts "account"
	"editor"
	"log"
	"mapping"
	"net/http"
	"orders"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

func main() {
	//dbquery.Connection("delivery", "order")

	callmap := func(w http.ResponseWriter, r *http.Request) {
		mapping.Pythonrun()
		mapping.Details(w, r)
	}
	createac := func(w http.ResponseWriter, r *http.Request) {
		accounts.Createaccount(w, r)
	}
	login := func(w http.ResponseWriter, r *http.Request) {

		accounts.Login(w, r)
	}
	edit := func(w http.ResponseWriter, r *http.Request) {

		accounts.Changepw(w, r)
	}
	load := func(w http.ResponseWriter, r *http.Request) {
		editor.Loadpage(w, r)

	}
	save := func(w http.ResponseWriter, r *http.Request) {
		editor.Savepage(w, r)

	}
	ac := func(w http.ResponseWriter, r *http.Request) {
		accounts.ACList(w, r)

	}
	order := func(w http.ResponseWriter, r *http.Request) {
		orders.OrderList(w, r)

	}
	editac := func(w http.ResponseWriter, r *http.Request) {
		accounts.Update(w, r)

	}
	editorder := func(w http.ResponseWriter, r *http.Request) {
		orders.Editorder(w, r)
	}
	dispatch := func(w http.ResponseWriter, r *http.Request) {
		orders.Dispatch(w, r)
	}
	callmaphandler := http.HandlerFunc(callmap)
	createachandler := http.HandlerFunc(createac)
	loginhandler := http.HandlerFunc(login)
	edithandler := http.HandlerFunc(edit)
	loadhandler := http.HandlerFunc(load)
	savehandler := http.HandlerFunc(save)
	orderhandler := http.HandlerFunc(order)
	achandler := http.HandlerFunc(ac)
	editachandler := http.HandlerFunc(editac)
	editorderhandler := http.HandlerFunc(editorder)
	dispatchhandler := http.HandlerFunc(dispatch)
	http.Handle("/", http.FileServer(http.Dir("../../../cms")))
	http.Handle("/callmap", callmaphandler)
	http.Handle("/create", createachandler)
	http.Handle("/login", loginhandler)
	http.Handle("/edit", edithandler)
	http.Handle("/webload", loadhandler)
	http.Handle("/websave", savehandler)
	http.Handle("/ordertable", orderhandler)
	http.Handle("/actable", achandler)
	http.Handle("/editac", editachandler)
	http.Handle("/dispatch", dispatchhandler)
	http.Handle("/editorder", editorderhandler)
	log.Fatal(http.ListenAndServe(":8081", nil))

}
