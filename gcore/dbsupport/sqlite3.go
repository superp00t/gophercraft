// +build sqlite3

package dbsupport

import (
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	Supported = append(Supported, "sqlite3")

	createFuncs[sqlite3] = func(url string) {
		sqlite3.Open()
	}
}
