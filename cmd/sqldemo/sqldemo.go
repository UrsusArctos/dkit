package main

import (
	"database/sql"
	"fmt"

	"github.com/UrsusArctos/dkit/pkg/aegisql"
)

const (
	dNameSQLite3 = "sqlite3"
	dNameMySQL   = "mysql"
	//
	queryCreate_SQLite3 = "CREATE TABLE IF NOT EXISTS [demotable] ([id] INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, [descr] VARCHAR(32) NULL);"
	queryCreate_MySQL   = "CREATE TABLE IF NOT EXISTS `demotable` (`id` INT(10) NOT NULL AUTO_INCREMENT, `descr` VARCHAR(32) NULL DEFAULT NULL, PRIMARY KEY (`id`));"
	queryInsertRow      = "INSERT INTO demotable (descr) VALUES ('new description');"
	querySelectData     = "SELECT * FROM demotable;"
)

var (
	driverName = [...]string{dNameSQLite3, dNameMySQL}
)

func main() {
	for dn := range driverName {
		var (
			driverDSN   string
			queryCreate string
		)
		fmt.Printf("%s: ", driverName[dn])
		// Compose DSN
		switch driverName[dn] {
		case dNameSQLite3:
			{
				driverDSN = aegisql.MakeDSN(driverName[dn], "/storage/shared/workbench/demo.s3db")
				queryCreate = queryCreate_SQLite3
			}
		case dNameMySQL:
			{
				driverDSN = aegisql.MakeDSN(driverName[dn], "dkituser", "dkitpass", "tcp", "10.0.0.10", "14030", "dkit")
				queryCreate = queryCreate_MySQL
			}
		}
		// Open database
		dbptr, err := sql.Open(driverName[dn], driverDSN)
		if err == nil {
			fmt.Print("opened,")
			AeSQL := aegisql.TAegiSQLDB{DB: dbptr}
			// Create table
			_, err2 := AeSQL.QuerySingle(queryCreate)
			if err2 == nil {
				fmt.Print("created,")
				// Insert some data
				_, err3 := AeSQL.QuerySingle(queryInsertRow)
				if err3 == nil {
					fmt.Print("inserted,")
					// Query some data
					dataRows, err4 := AeSQL.QueryData(querySelectData)
					if err4 == nil {
						fmt.Print("queried,")
						// Ouput received data
						for {
							rowData := dataRows.UnloadNextRow()
							if rowData == nil {
								break
							}
							fmt.Printf("data:(%+v)", rowData)
						}
					} else {
						fmt.Printf("DB data query error %+v\n", err4)
						break
					}
				} else {
					fmt.Printf("DB insert query error %+v\n", err3)
					break
				}
			} else {
				fmt.Printf("DB single query error %+v\n", err2)
				break
			}
		} else {
			fmt.Printf("DB open error %+v\n", err)
			break
		}
		dbptr.Close()
		fmt.Println()
	}
}
