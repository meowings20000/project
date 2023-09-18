package dbquery

import (
	"fmt"

	rethinkdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

type RdbSess struct {
	session   *rethinkdb.Session
	dbName    string
	tableName string
}

func (r *RdbSess) Create(dbName string, tablename string) error {
	_, err := rethinkdb.DBCreate(dbName).Run(r.session)
	_, err = rethinkdb.DB(dbName).TableCreate(tablename).Run(r.session)

	return err
}

func Connectdb(dbname string, tablename string) (RdbSess, error) {
	var rdbSession RdbSess
	var message string
	var dresult, tresult bool

	session, err := rethinkdb.Connect(rethinkdb.ConnectOpts{
		Address: "localhost:28015",
	})

	if err != nil {
		return rdbSession, err
	}

	rdbSession.session = session

	dexists, err := rethinkdb.DBList().Contains(dbname).Run(session)
	texists, err := rethinkdb.DB(dbname).TableList().Contains(tablename).Run(session)
	if err != nil {
		return rdbSession, err
	}

	err = dexists.One(&dresult)
	err = texists.One(&tresult)
	if dresult && tresult {
		message = "Database and table already exist"
		fmt.Println(message)

	} else {
		message = "Database and table created"
		err = rdbSession.Create(dbname, tablename)

		fmt.Println(message)
	}

	rdbSession.dbName = dbname
	rdbSession.tableName = tablename
	if err != nil {
		return rdbSession, err
	}

	return rdbSession, nil
}

func (r *RdbSess) Count() (*rethinkdb.Cursor, error) {
	num, err := rethinkdb.DB(r.dbName).Table(r.tableName).Count().Run(r.session)

	if err != nil {
		return num, err
	}
	return num, err
}

func (r *RdbSess) Insert(data interface{}) error {
	_, err := rethinkdb.DB(r.dbName).Table(r.tableName).Insert(data).RunWrite(r.session)
	if err != nil {
		return err
	}

	return nil
}

func (r *RdbSess) Query(query rethinkdb.Term) (*rethinkdb.Cursor, error) {
	cursor, err := query.Run(r.session)

	if err != nil {
		panic(err)
	}
	return cursor, nil
}

func (r *RdbSess) Closeconn() error {
	err := r.session.Close()
	if err != nil {
		return err
	}

	return nil
}
