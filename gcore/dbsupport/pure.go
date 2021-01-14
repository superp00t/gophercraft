package dbsupport

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/superp00t/etc/yo"
)

var Supported = []string{
	"mysql",
}

var PathFormat = map[string]string{
	"mysql": "user:password@/databaseName",
}

func Create(driver, url string) error {
	cFunc, ok := createFuncs[driver]
	if !ok {
		return fmt.Errorf("no create func for %s", driver)
	}

	return cFunc(url)
}

var createFuncs = map[string]func(url string) error{
	"mysql": func(url string) error {

		urlp := strings.Split(url, "/")
		if len(urlp) != 2 {
			return fmt.Errorf("dbsupport: mysql needs a database name user:pass@host/databasename")
		}

		dbURL := urlp[0] + "/"

		db, err := sql.Open("mysql", dbURL)
		if err != nil {
			return err
		}

		dbName := urlp[1]

		matched, err := regexp.MatchString("^[a-zA-Z0-9_]*$", dbName)
		if err != nil {
			panic(err)
		}

		if !matched {
			return fmt.Errorf("invalid database name %s", dbName)
		}

		yo.Ok("Creating", dbName)

		if _, err := db.Exec("CREATE DATABASE " + dbName); err != nil {
			yo.Warn(err)
		}
		return nil
	},
}
