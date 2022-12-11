package sdb

import (
	"database/sql"
	"errors"
	"reflect"

	"github.com/DeniesKresna/gohelper/utinterface"
	"github.com/DeniesKresna/gohelper/utlog"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type DBInstance struct {
	DB  *sqlx.DB
	Tx  *sqlx.Tx
	Tag string
}

// Init DB. sqlx by jmoiron is used in this package
//
// dbDriver is the driver. exp: mysql
//
// dbTag is the struct field identifier to scan rows value
// exp: name string `db:"name"`
// on above example you should pass "db" to this param
//
// dsn is the connection config.
// exp: user:password!@tcp(localhost:3306)/dbName?parseTime=true
func InitDB(dbDriver string, dsn string) (dbIns *DBInstance, err error) {
	db, err := sqlx.Connect(dbDriver, dsn)
	if err != nil {
		utlog.Errorf("Error while get db, err: %+v", err)
	}

	dbTag := "db"

	dbIns = &DBInstance{
		DB:  db,
		Tag: dbTag,
	}

	utlog.Info("InitDB done")
	return
}

// Execute Create Update Delete query
//
// query is the sql statement
//
// args is the argument to be passed to the query
func (d *DBInstance) Exec(query string, args ...any) (sql.Result, error) {
	if d.Tx != nil {
		return d.Tx.Exec(query, args...)
	}
	return d.DB.Exec(query, args...)
}

// Get only one data row from sql
//
// dest is destination of the struct. should be pointer
//
// query is the sql statement
//
// args is the argument to be passed to the query
func (d *DBInstance) Take(dest interface{}, query string, args ...interface{}) error {
	var (
		resp *sqlx.Rows
		err  error
	)

	if !utinterface.IsPointer(dest) {
		return errors.New("destination should be pointer")
	}

	if utinterface.IsPointerOfStruct(dest) {
		if d.Tx != nil {
			resp, err = d.Tx.Queryx(query, args...)
		} else {
			resp, err = d.DB.Queryx(query, args...)
		}
	} else {
		if d.Tx != nil {
			err = d.Tx.Get(dest, query, args...)
		} else {
			err = d.DB.Get(dest, query, args...)
		}
		return nil
	}

	if err != nil {
		return err
	}

	count := 0
	for resp.Next() {
		err = resp.StructScan(dest)
		if err != nil {
			return err
		}
		count++
		break
	}

	if count <= 0 {
		return errors.New("No data found")
	}

	return nil
}

// select some data rows from sql
//
// dest is destination of the struct. should be pointer
//
// query is the sql statement
//
// args is the argument to be passed to the query
func (d *DBInstance) Select(dest interface{}, query string, args ...interface{}) error {
	var (
		resp *sqlx.Rows
		err  error
	)

	if !utinterface.IsPointerOfSliceOfStruct(dest) {
		return errors.New("dest should be slice of pointer of struct")
	}

	reflectDestType := reflect.TypeOf(dest).Elem().Elem()
	reflectDestValue := reflect.ValueOf(dest).Elem()

	if d.Tx != nil {
		resp, err = d.Tx.Queryx(query, args...)
	} else {
		resp, err = d.DB.Queryx(query, args...)
	}
	if err != nil {
		return err
	}

	for resp.Next() {
		intPtr := reflect.New(reflectDestType)

		err = resp.StructScan(intPtr.Interface())
		if err != nil {
			return err
		}

		reflectDestValue = reflect.Append(reflectDestValue, intPtr.Elem())
	}
	reflect.ValueOf(dest).Elem().Set(reflectDestValue)

	return nil
}

// Get only one data row from sql with original sqlx Get
//
// dest is destination of the struct. should be pointer
//
// query is the sql statement
//
// args is the argument to be passed to the query
func (d *DBInstance) Get(dest interface{}, query string, args ...interface{}) error {
	var (
		err error
	)

	if !utinterface.IsPointer(dest) {
		return errors.New("destination should be pointer")
	}

	if d.Tx != nil {
		err = d.Tx.Get(dest, query, args...)
	} else {
		err = d.DB.Get(dest, query, args...)
	}

	return err
}

// Select some data rows from sql with original sqlx Queryx
//
// dest is destination of the struct. should be pointer
//
// query is the sql statement
//
// args is the argument to be passed to the query
func (d *DBInstance) Queryx(dest interface{}, query string, args ...interface{}) error {
	var (
		resp *sqlx.Rows
		err  error
	)

	if !utinterface.IsPointerOfSliceOfStruct(dest) {
		return errors.New("dest should be slice of pointer of struct")
	}

	reflectDestType := reflect.TypeOf(dest).Elem().Elem()
	reflectDestValue := reflect.ValueOf(dest).Elem()

	if d.Tx != nil {
		resp, err = d.Tx.Queryx(query, args...)
	} else {
		resp, err = d.DB.Queryx(query, args...)
	}
	if err != nil {
		return err
	}

	for resp.Next() {
		intPtr := reflect.New(reflectDestType)

		err = resp.StructScan(intPtr.Interface())
		if err != nil {
			return err
		}

		reflectDestValue = reflect.Append(reflectDestValue, intPtr.Elem())
	}
	reflect.ValueOf(dest).Elem().Set(reflectDestValue)

	return err
}

// Start DB Transaction
func (d *DBInstance) StartTx() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Recovered in StartTx")
		}
	}()

	if d.Tx != nil {
		return
	}

	d.Tx = d.DB.MustBegin()
	return
}

// Commit DB Transaction
func (d *DBInstance) Commit() (err error) {
	if d.Tx != nil {
		err = d.Tx.Commit()
		if d.Tx != nil {
			d.Tx = nil
		}
	}
	return
}

// Cancel DB Transaction
func (d *DBInstance) Rollback() (err error) {
	if d.Tx != nil {
		err = d.Tx.Rollback()
		if d.Tx != nil {
			d.Tx = nil
		}
	}
	return
}

// Submit DB Transaction. Its canceled DB transaction on error
func (d *DBInstance) SubmitTx(err error) error {
	if err != nil {
		return d.Rollback()
	}
	return d.Commit()
}
