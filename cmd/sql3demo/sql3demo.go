package main

import (
	"fmt"

	"github.com/UrsusArctos/dkit/pkg/sqlite"
)

const (
	CreateTableSQL = "CREATE TABLE IF NOT EXISTS [demotable] " +
		"([id] INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, [descr] VARCHAR(32) NULL)"
	InsertRowSQL = "INSERT INTO demotable (descr) VALUES ('new description')"
	SelectSQL    = "SELECT * FROM demotable WHERE id>1"
)

func main() {
	db, err := sqlite.ConnectToDB("/storage/shared/workbench/demo.s3db")
	fmt.Printf("Connect: %+v\n", err)
	if err == nil {
		defer db.Close()
		rrows, serr := db.QueryData(SelectSQL)
		if serr == nil {
			for {
				rdata := rrows.UnloadNextRow()
				if rdata == nil {
					break
				}
				fmt.Printf("Select: %+v\n", rdata)
			}
		}
	}
	//	_, err = db.QuerySingle(CreateTableSQL)
	//	fmt.Printf("Create table: %+v\n", err)
	//	sr, serr := db.QuerySingle(InsertRowSQL)
	//	lid, _ := sr.LastInsertId()
	//	fmt.Printf("Insert row: as %d, %+v\n", lid, serr)

}
