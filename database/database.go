package database

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

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
func InitDB(dbDriver string, dbTag string, dsn string) (dbIns *DBInstance, err error) {
	db, err := sqlx.Connect(dbDriver, dsn)
	if err != nil {
		utlog.Errorf("Error while get db, err: %+v", err)
	}

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
func (d *DBInstance) Get(dest interface{}, query string, args ...interface{}) error {
	var (
		resp    *sqlx.Rows
		respMap = make(map[string]interface{})
		err     error
	)

	if !utinterface.IsPointerOfStruct(dest) {
		return errors.New("destination should be pointer of struct")
	}

	if d.Tx != nil {
		resp, err = d.Tx.Queryx(query, args...)
	} else {
		resp, err = d.DB.Queryx(query, args...)
	}
	if err != nil {
		return err
	}

	for resp.Next() {
		err = resp.MapScan(respMap)
		if err != nil {
			return err
		}
		break
	}

	if len(respMap) <= 0 {
		return errors.New("No data found")
	}

	err = d.ScanToStruct(respMap, dest)
	if err != nil {
		return err
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
		resp    *sqlx.Rows
		respMap = make(map[string]interface{})
		err     error
	)

	if !utinterface.IsPointerOfSliceOfStruct(dest) {
		return errors.New("dest should be slice of pointer of struct")
	}

	reflectDestType := reflect.TypeOf(dest).Elem().Elem()
	reflectDestValue := reflect.ValueOf(dest)

	if d.Tx != nil {
		resp, err = d.Tx.Queryx(query, args...)
	} else {
		resp, err = d.DB.Queryx(query, args...)
	}
	if err != nil {
		return err
	}

	for resp.Next() {
		err = resp.MapScan(respMap)
		if err != nil {
			return err
		}

		intPtr := reflect.New(reflectDestType)
		err = d.ScanToStruct(respMap, intPtr)
		if err != nil {
			return err
		}

		reflectDestValue = reflect.Append(reflectDestValue, intPtr)
	}

	return nil
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

func (d *DBInstance) ScanToStruct(ins map[string]interface{}, destination interface{}) (errs error) {
	// It's possible we can cache this, which is why precompute all these ahead of time.
	getDBTagValues := func(t reflect.StructTag) string {
		if jt, ok := t.Lookup(d.Tag); ok {
			tagValue := strings.TrimSpace(strings.Split(jt, ",")[0])
			if tagValue != "-" {
				return tagValue
			}
		}
		return ""
	}

	convertInterfaceToStruct := func(in map[string]interface{}, dest interface{}, getDBTagValue func(t reflect.StructTag) string) (err error) {
		reflectValue := reflect.ValueOf(dest).Elem()
		reflectType := reflectValue.Type()

		for i := 0; i < reflectValue.NumField(); i++ {
			typeField := reflectType.Field(i)
			tag := typeField.Tag
			dbKey := getDBTagValue(tag)
			if dbKey == "" {
				continue
			}

			dbVal, ok := in[dbKey]
			if !ok || dbVal == nil {
				continue
			}

			f := reflectValue.FieldByIndex(typeField.Index)
			fieldValue := f.Interface()

			if f.IsValid() {
				if f.CanSet() {
					switch fieldValue.(type) {
					case int, int64:
						strVal := fmt.Sprintf("%d", dbVal)
						realVal, err := strconv.Atoi(strVal)
						if err != nil {
							return err
						}
						f.SetInt(int64(realVal))
						continue
					case string:
						strVal := fmt.Sprintf("%s", dbVal)
						f.SetString(strVal)
						continue
					case bool:
						strVal := fmt.Sprintf("%d", dbVal)
						var realVal bool
						if strVal == "1" {
							realVal = true
						} else {
							realVal = false
						}
						f.SetBool(realVal)
						continue
					case time.Time:
						strVal := fmt.Sprintf("%s", dbVal)
						tm, err := time.Parse("2006-01-02 15:04:05 Z0700 MST", strVal)
						if err != nil {
							return err
						}
						f.Set(reflect.ValueOf(tm))
						continue
					}
				}
			}
		}
		return nil
	}

	return convertInterfaceToStruct(ins, destination, getDBTagValues)
}
