package aegisql

import (
	"fmt"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type (
	TAegiSQLDB struct {
		*sql.DB
	}

	TAegiSQLRows struct {
		*sql.Rows
	}

	TAegiSQLDataRow = map[string]string
)

func MakeDSN(driverName string, arg ...string) string {
	switch driverName {
	case "sqlite3":
		// arg[0] 	- DB file name
		return fmt.Sprintf("file:%s?cache=shared&mode=rwc", arg[0])
	case "mysql":
		// arg[0]	- username
		// arg[1]	- password
		// arg[2]	- protocol (see net.Dial)
		// arg[3]	- host address
		// arg[4]	- port
		// arg[5]	- database name
		return fmt.Sprintf("%s:%s@%s(%s:%s)/%s", arg[0], arg[1], arg[2], arg[3], arg[4], arg[5])
	default:
		return ""
	}
}

func (sqldb TAegiSQLDB) QuerySingle(query string, args ...any) (sql.Result, error) {
	stmt, perr := sqldb.Prepare(query)
	if perr == nil {
		defer stmt.Close()
		res, eerr := stmt.Exec(args...)
		if eerr == nil {
			return res, eerr
		}
		return nil, eerr
	}
	return nil, perr
}

func (sqldb TAegiSQLDB) QueryData(query string, args ...any) (TAegiSQLRows, error) {
	rawRows, err := sqldb.Query(query, args...)
	if err == nil {
		return TAegiSQLRows{rawRows}, nil
	}
	return TAegiSQLRows{}, err
}

func (rows TAegiSQLRows) UnloadNextRow() TAegiSQLDataRow {
	if rows.Next() {
		columns, cerr := rows.Columns()
		if cerr == nil {
			pointers := make([]any, len(columns))
			values := make([]string, len(columns))
			for i := range pointers {
				pointers[i] = &values[i]
			}
			scanerr := rows.Scan(pointers...)
			if scanerr == nil {
				result := make(TAegiSQLDataRow)
				for i := range values {
					result[columns[i]] = values[i]
				}
				return result
			}
		}
	}
	return nil
}
