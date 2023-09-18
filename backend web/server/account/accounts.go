package accounts

import (
	"encoding/json"
	"fmt"
	"net/http"
	dbquery "rethinkdb"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/rethinkdb/rethinkdb-go.v6"
)

type jsonEmployee struct {
	StaffId int    `json:"StaffId" rethinkdb:"StaffId,omitempty"`
	Name    string `json:"Name" rethinkdb:"Name,omitempty"`
	Email   string `json:"Email" rethinkdb:"Email,omitempty"`
	Role    string `json:"Role" rethinkdb:"Role,omitempty"`
}
type Employee struct {
	StaffId  int    `json:"StaffId" rethinkdb:"StaffId,omitempty"`
	Name     string `json:"Name" rethinkdb:"Name,omitempty"`
	Email    string `json:"Email" rethinkdb:"Email,omitempty"`
	Role     string `json:"Role" rethinkdb:"Role,omitempty"`
	Password string `json:"Password" rethinkdb:"Password,omitempty"`
}

type respond struct {
	Status   string   `json:"Status"`
	Employee Employee `json:"Employee"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Createaccount(w http.ResponseWriter, r *http.Request) {
	var staffid int
	var repeated bool
	r.ParseForm()
	stringpw := strings.Join(r.Form["password"], " ")
	hashpw, _ := HashPassword(stringpw)
	session, _ := dbquery.Connectdb("user", "users")
	query := rethinkdb.DB("user").Table("users").Filter(rethinkdb.Row.Field("Email").Eq(strings.Join(r.Form["email"], " "))).Distinct().Count()
	result, _ := session.Query(query)
	_ = result.One(&repeated)
	if !repeated {
		query = rethinkdb.DB("user").Table("users").Count()
		result, _ := session.Query(query)
		_ = result.One(&staffid)
		data := Employee{staffid + 1, strings.Join(r.Form["name"], " "), strings.Join(r.Form["email"], " "), strings.Join(r.Form["role"], " "), hashpw}
		session.Insert(data)
		w.WriteHeader(200)
	} else {
		w.WriteHeader(420)
		w.Write([]byte("You already have an account"))
	}
	session.Closeconn()
}

func Update(w http.ResponseWriter, r *http.Request) {
	//var staffid int
	var repeated bool
	r.ParseForm()
	stringpw := strings.Join(r.Form["password"], " ")
	hashpw, _ := HashPassword(stringpw)
	session, _ := dbquery.Connectdb("user", "users")
	query := rethinkdb.DB("user").Table("users").Filter(rethinkdb.Row.Field("Email").Eq(strings.Join(r.Form["oldemail"], " "))).Distinct().Count()
	result, _ := session.Query(query)
	_ = result.One(&repeated)
	if repeated {
		//data := Employee{staffid, strings.Join(r.Form["name"], " "), strings.Join(r.Form["email"], " "), strings.Join(r.Form["role"], " "), hashpw}
		query = rethinkdb.DB("user").Table("users").Filter(map[string]string{"Email": strings.Join(r.Form["oldemail"], " ")}).Update(map[string]string{"Email": strings.Join(r.Form["email"], " "), "Role": strings.Join(r.Form["role"], " "), "Name": strings.Join(r.Form["name"], " "), "Password": hashpw})
		session.Query(query)
		w.WriteHeader(200)
	} else {
		w.WriteHeader(420)
		w.Write([]byte("Please create an account"))
	}
	session.Closeconn()
}

func Login(w http.ResponseWriter, r *http.Request) {
	var employee Employee
	r.ParseForm()

	session, _ := dbquery.Connectdb("user", "users")
	stringemail := strings.Join(r.Form["email"], " ")
	stringpw := strings.Join(r.Form["password"], " ")
	query := rethinkdb.DB("user").Table("users").Filter(map[string]string{"Email": stringemail})
	result, _ := session.Query(query)
	_ = result.One(&employee)
	check := CheckPasswordHash(employee.Password, stringpw)
	if check {
		w.Header().Set("Content-Type", "application/json")
		repond := respond{Status: "success", Employee: employee}
		jsonBytes, _ := json.Marshal(repond)
		w.Write(jsonBytes)
	} else {
		w.Header().Set("Content-Type", "application/json")
		repond := respond{Status: "fail", Employee: Employee{0, "null", "null", "null", "null"}}
		jsonBytes, _ := json.Marshal(repond)
		w.Write(jsonBytes)
	}
	session.Closeconn()
}

func Changepw(w http.ResponseWriter, r *http.Request) {
	var employee Employee
	r.ParseForm()
	fmt.Println("email:", r.Form["email"])
	fmt.Println("password:", r.Form["Old"])
	session, _ := dbquery.Connectdb("user", "users")
	stringemail := strings.Join(r.Form["email"], " ")
	stringemail = stringemail[9:]
	stringemail = strings.TrimSuffix(stringemail, "\"")

	oldpw := strings.Join(r.Form["old"], "")
	newpw := strings.Join(r.Form["new"], "")
	confirm := strings.Join(r.Form["confirm"], "")
	query := rethinkdb.DB("user").Table("users").Filter(map[string]string{"Email": stringemail})
	result, _ := session.Query(query)
	_ = result.One(&employee)

	check := CheckPasswordHash(employee.Password, oldpw)
	if check {
		if newpw == confirm {
			hashpw, _ := HashPassword(newpw)
			query := rethinkdb.DB("user").Table("users").Filter(map[string]string{"Email": stringemail}).Update(map[string]interface{}{"Password": hashpw})
			result, _ := session.Query(query)
			fmt.Print(result)
			w.Header().Set("Content-Type", "application/json")
			jsonBytes, _ := json.Marshal("Success ,You have changed to your new password")
			w.Write(jsonBytes)
		} else {
			w.Header().Set("Content-Type", "application/json")
			jsonBytes, _ := json.Marshal("Error ,Your new password does not match the confirm password")
			w.Write(jsonBytes)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		jsonBytes, _ := json.Marshal("Error ,Your old password is not correct")
		w.Write(jsonBytes)
	}
	session.Closeconn()
}
func ACList(w http.ResponseWriter, r *http.Request) {
	var employee []jsonEmployee
	session, _ := dbquery.Connectdb("user", "users")
	query := rethinkdb.DB("user").Table("users")
	result, _ := session.Query(query)
	_ = result.All(&employee)
	jsonBytes, _ := json.Marshal(employee)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
	session.Closeconn()
}
