package gosrc

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"github.com/alexbrainman/odbc/api"
)

// func (c *Conn) FetchColumns(m map[string][]string) (map[string][]string, error) {
func (c *Conn) FetchColumns(tableName string) (string, error) {
	if tableName == "" || tableName == " " {
		return "", errors.New("empty input string")
	}
	// for k := range m {
	var out api.SQLHANDLE
	ret := api.SQLAllocHandle(api.SQL_HANDLE_STMT, api.SQLHANDLE(c.h), &out)
	if IsError(ret) {
		return "", c.newError("SQLAllocHandle", c.h)
	}
	stmt := api.SQLHSTMT(out)
	err := Drv.Stats.updateHandleCount(api.SQL_HANDLE_STMT, 1)
	fmt.Println("column statement allocated: ", Drv.Stats)
	if err != nil {
		return "", err
	}
	start := time.Now()
	// for k := range m {
	// cn := api.StringToUTF16(k)
	cn := api.StringToUTF16(tableName)
	ret = api.SQLColumns(stmt, nil, 0, nil, 0, (*api.SQLWCHAR)(unsafe.Pointer(&cn[0])), api.SQL_NTS, nil, 0)
	if IsError(ret) {
		return "", c.newError("SQLColumns", stmt)
	}
	var colname [128]api.SQLCHAR
	var ind api.SQLLEN
	var colList []string
	// api.SQLBindCol(stmt, 4, api.SQL_C_CHAR, api.SQLPOINTER(&colname), 128, &ind)
	ret = api.SQLBindCol(stmt, 4, api.SQL_CHAR, api.SQLPOINTER(&colname), api.SQLLEN(len(colname)), &ind)
	if IsError(ret) {
		return "", NewError("SQLBindCol", stmt)
	}
	for {
		ret = api.SQLFetch(stmt)
		if IsError(ret) {
			fmt.Println("time taken: ", time.Since(start))
			releaseHandle(stmt)
			if colList == nil {
				return "", errors.New("colList is null")
			}
			j, err := json.Marshal(colList)
			if err != nil {
				return "", err
			}
			return string(j), nil
			// break
		}
		colstring := []byte{}
		for _, chardata := range colname {
			if chardata != 0 {
				colstring = append(colstring, byte(chardata))
			}
			var newcolname [128]api.SQLCHAR
			colname = newcolname
		}
		// m[k] = append(m[k], string(colstring))
		colList = append(colList, string(colstring))
		// log.Printf(" Type Cols : " + string(m))

	}
	// }
	fmt.Println("columns stmt released: ", Drv.Stats)
	return "", nil
}

func (c *Conn) FetchTables() (string, error) {
	m := make(map[string][]string)
	// fmt.Println(m)
	var out api.SQLHANDLE
	ret := api.SQLAllocHandle(api.SQL_HANDLE_STMT, api.SQLHANDLE(c.h), &out)
	if IsError(ret) {
		return "", c.newError("SQLAllocHandle", c.h)
	}
	h := api.SQLHSTMT(out)
	err := Drv.Stats.updateHandleCount(api.SQL_HANDLE_STMT, 1)
	if err != nil {
		return "", err
	}
	fmt.Println("table statement allocated: ", Drv.Stats)
	tn := api.StringToUTF16("TABLE")
	tne := (*api.SQLWCHAR)(unsafe.Pointer(&tn[0]))
	// b := api.StringToUTF16("TABLE")
	ret = api.SQLTables(h, nil, 0, nil, 0, nil, 0, tne, api.SQL_NTS)
	// ret = api.SQLTables(s.h, nil, 0, nil, 0, nil, 0, nil, 0)
	if IsError(ret) {
		defer releaseHandle(h)
		return "SQLTables fail", c.newError("SQLTables", h)
	}
	for {
		var len api.SQLLEN
		var Buffer [128]api.SQLCHAR
		ret = api.SQLFetch(h)
		if !sql_succeeded(ret) {
			releaseHandle(h)

			// c.FetchColumns(m)
			j, err := json.Marshal(m)
			if err != nil {
				return "", err
			}
			fmt.Println(string(j))
			return string(j), nil
		}
		ret = api.SQLGetData(h, 3, api.SQL_C_CHAR, api.SQLPOINTER(&Buffer), 128, &len)
		if IsError(ret) {
			return "SQLGetData", c.newError("SQLGetData", h)
		}
		typestring := []byte{}
		for _, chardata := range Buffer {
			if chardata != 0 {
				typestring = append(typestring, byte(chardata))
			}
		}
		// m[string(typestring)] = nil
		m[string(typestring)] = append(m[string(typestring)], "beHumble")
	}
	fmt.Println("tables stmt: ", Drv.Stats)
	return "tablesjson", nil
}

func sql_succeeded(ret api.SQLRETURN) bool {
	return (uint32(ret) & (^uint32(1))) == 0
}

// func (s *Statement) Close() (err error) {
// 	h := s.h
// 	defer func() {
// 		s.h = api.SQLHSTMT(api.SQL_NULL_HSTMT)
// 		e := releaseHandle(h)
// 		if err == nil {
// 			err = e
// 		}
// 	}()
// 	return err
// }
