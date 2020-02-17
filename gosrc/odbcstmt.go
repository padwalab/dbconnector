package gosrc

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"unsafe"

	"github.com/alexbrainman/odbc/api"
)

type Value interface{}

//ODBCStmt the container for all the data to be displayed
type ODBCStmt struct {
	h          api.SQLHSTMT
	Parameters []Parameter
	Cols       []Column
	mu         sync.Mutex
	usedByStmt bool
	usedByRows bool
}

//ResultSet contains the result set metadata
type ResultSet struct {
	ColumnName string `json:"columnName"`
	ColumnDT   string `json:"columnDataType"`
}

//ParamSet contains params metadata
type ParamSet struct {
	ParamName string `json:"paramName"`
	ParamDT   string `json:"paramDataType"`
	Nullable  bool   `json:"nullable"`
}

//StmtMetadata container for metadata of prepared statement
type StmtMetadata struct {
	Query   string      `json:"query"`
	Results []ResultSet `bson:"results" json:"resultSet"`
	Params  []ParamSet  `bson:"params" json:"parameters"`
}

type MasterResultSet struct {
	Cols       []string
	ResultData []interface{}
}

func describeParam(h api.SQLHSTMT, idx int) (sqltype api.SQLSMALLINT, size api.SQLULEN, nullable api.SQLSMALLINT, ret api.SQLRETURN) {
	var decimal api.SQLSMALLINT
	ret = api.SQLDescribeParam(h, api.SQLUSMALLINT(idx+1),
		&sqltype, &size, &decimal, &nullable)
	return sqltype, size, nullable, ret
}

func getSqlType(sqltype api.SQLSMALLINT) string {
	switch sqltype {
	case api.SQL_BIT:
		return "SQL_BIT"
	case api.SQL_TINYINT:
		return "number"
	case api.SQL_SMALLINT:
		return "number"
	case api.SQL_INTEGER:
		return "number"
	case api.SQL_BIGINT:
		return "number"
	case api.SQL_NUMERIC, api.SQL_DECIMAL, api.SQL_FLOAT, api.SQL_REAL, api.SQL_DOUBLE:
		return "number"
	case api.SQL_TYPE_TIMESTAMP:
		return "SQL_TYPE_TIMESTAMP"
	case api.SQL_TYPE_DATE:
		return "SQL_TYPE_DATE"
	case api.SQL_TYPE_TIME:
		return "SQL_TYPE_TIME"
	case api.SQL_SS_TIME2:
		return "api.SQL_SS_TIME2"
	case api.SQL_GUID:
		return "SQL_GUID"
	case api.SQL_CHAR, api.SQL_VARCHAR:
		return "string"
	case api.SQL_WCHAR, api.SQL_WVARCHAR:
		return "string"
	case api.SQL_BINARY, api.SQL_VARBINARY:
		return "string"
	case api.SQL_LONGVARCHAR:
		return "string"
	case api.SQL_WLONGVARCHAR, api.SQL_SS_XML:
		return "string"
	case api.SQL_LONGVARBINARY:
		return "SQL_LONGVARBINARY"
	default:
		return ""
	}
}

// PrepareODBCStmt creates the intended metadata set as prepareResponse.json
func (c *Conn) PrepareODBCStmt(query string) (*ODBCStmt, StmtMetadata, error) {
	var out api.SQLHANDLE
	var stmtMetadata StmtMetadata
	ret := api.SQLAllocHandle(api.SQL_HANDLE_STMT, api.SQLHANDLE(c.h), &out)
	if IsError(ret) {
		return nil, stmtMetadata, c.newError("SQLAllocHandle", c.h)
	}
	h := api.SQLHSTMT(out)
	err := Drv.Stats.updateHandleCount(api.SQL_HANDLE_STMT, 1)
	if err != nil {
		return nil, stmtMetadata, err
	}
	stmtMetadata.Query = query
	b := api.StringToUTF16(query)
	ret = api.SQLPrepare(h, (*api.SQLWCHAR)(unsafe.Pointer(&b[0])), api.SQL_NTS)
	if IsError(ret) {
		defer releaseHandle(h)
		return nil, stmtMetadata, c.newError("SQLPrepare", h)
	}
	var nCols api.SQLSMALLINT
	ret = api.SQLNumResultCols(h, &nCols)
	if IsError(ret) {
		return nil, stmtMetadata, NewError("SQLNumResultCols", h)
	}
	fmt.Println("number of cols in result set: ", nCols)
	for i := 0; i < int(nCols); i++ {
		var result ResultSet
		namebuf := make([]uint16, 150)
		namelen, sqltype, size, ret := describeColumn(h, i, namebuf)
		if ret == api.SQL_SUCCESS_WITH_INFO && namelen > len(namebuf) {
			// try again with bigger buffer
			namebuf = make([]uint16, namelen)
			namelen, sqltype, size, ret = describeColumn(h, i, namebuf)
		}
		if IsError(ret) {
			return nil, stmtMetadata, NewError("SQLDescribeCol", h)
		}
		if namelen > len(namebuf) {
			// still complaining about buffer size
			return nil, stmtMetadata, errors.New("Failed to allocate column name buffer")
		}
		result.ColumnName = api.UTF16ToString(namebuf[:namelen])
		result.ColumnDT = getSqlType(sqltype)
		stmtMetadata.Results = append(stmtMetadata.Results, result)
		fmt.Println("Name is ", result.ColumnName, "sqltype: ", result.ColumnDT, "size: ", size)
	}
	var nParams api.SQLSMALLINT
	ret = api.SQLNumParams(h, &nParams)
	if IsError(ret) {
		return nil, stmtMetadata, NewError("SQLNumParams", h)
	}
	fmt.Println("number of Params in result set: ", nParams)
	for i := 0; i < int(nParams); i++ {
		var param ParamSet
		sqltype, size, nullable, ret := describeParam(h, i)
		if IsError(ret) {
			return nil, stmtMetadata, NewError("SQLDescribeParam", h)
		}
		param.ParamName = fmt.Sprintf("Param[%d]", i+1)
		param.ParamDT = getSqlType(sqltype)
		param.Nullable = !(int(nullable) == 0)
		stmtMetadata.Params = append(stmtMetadata.Params, param)
		fmt.Println("Param number: ", i, " DataType", param.ParamDT, "size: ", size, "Nullable: ", param.Nullable)
	}
	ps, err := ExtractParameters(h)
	if err != nil {
		defer releaseHandle(h)
		return nil, stmtMetadata, err
	}
	// for _, item := range ps {
	// 	fmt.Println("Data from Parameters; ", getSqlType(item.SQLType))
	// }

	return &ODBCStmt{h: h, Parameters: ps}, stmtMetadata, nil
}

func (s *ODBCStmt) Query(args []interface{}) (*Rows, error) {
	err := s.BindColumns()
	if err != nil {
		defer releaseHandle(s.h)
		return nil, err
	}
	return &Rows{os: s}, nil
}

func (s *ODBCStmt) Exec(args []interface{}) (*MasterResultSet, error) {
	// fmt.Println("args through the eyes of Exec ", args)
	var mrs MasterResultSet
	if len(args) == 0 {
		fmt.Println("no params to bind")
	}
	if len(args) != len(s.Parameters) {
		return nil, fmt.Errorf("wrong number of arguments %d, %d expected", len(args), len(s.Parameters))
	}
	for i, a := range args {
		if err := s.Parameters[i].BindValue(s.h, i, a); err != nil {
			return nil, err
		}
	}
	ret := api.SQLExecute(s.h)
	if ret == api.SQL_NO_DATA {
		// success but no data to report
		fmt.Println("success but no data to return")
		return nil, nil
	}
	if IsError(ret) {
		defer releaseHandle(s.h)
		return nil, NewError("SQLExecute", s.h)
	}
	a, err := s.Query(args)
	// fmt.Println("call to Query")
	if err != nil {
		return nil, err
	}
	// resultSetLen := len(a.Columns())
	mrs.Cols = a.Columns()

	for {
		rset := make([]interface{}, len(a.Columns()))
		err := a.Next(rset)
		// fmt.Println("call to Next")
		if err == io.EOF {
			a.Close()
			s.releaseHandle()
			break
		}
		for i, rs := range rset {
			switch rs.(type) {
			default:
				// fmt.Print(rs, "\t")
				rset[i] = rs
			case []uint8:
				// fmt.Print(string(rs.([]uint8)), "\t")
				rset[i] = string(rs.([]uint8))
			}
		}
		mrs.ResultData = append(mrs.ResultData, rset)
		// fmt.Print("\n", mrs.ResultData)
	}
	return &mrs, nil
}

func (s *ODBCStmt) closeByRows() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.usedByRows {
		defer func() { s.usedByRows = false }()
		if s.usedByStmt {
			ret := api.SQLCloseCursor(s.h)
			if IsError(ret) {
				return NewError("SQLCloseCursor", s.h)
			}
			return nil
		} else {
			return s.releaseHandle()
		}
	}
	return nil
}

func (s *ODBCStmt) closeByStmt() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.usedByStmt {
		defer func() { s.usedByStmt = false }()
		if !s.usedByRows {
			return s.releaseHandle()
		}
	}
	return nil
}

func (s *ODBCStmt) releaseHandle() error {
	h := s.h
	s.h = api.SQLHSTMT(api.SQL_NULL_HSTMT)
	return releaseHandle(h)
}

func (s *ODBCStmt) BindColumns() error {
	// count columns
	var n api.SQLSMALLINT
	ret := api.SQLNumResultCols(s.h, &n)
	if IsError(ret) {
		return NewError("SQLNumResultCols", s.h)
	}
	if n < 1 {
		return errors.New("Stmt did not create a result set")
	}
	// fetch column descriptions
	s.Cols = make([]Column, n)
	binding := true
	for i := range s.Cols {
		c, err := NewColumn(s.h, i)
		if err != nil {
			return err
		}
		s.Cols[i] = c
		// Once we found one non-bindable column, we will not bind the rest.
		// http://www.easysoft.com/developer/languages/c/odbc-tutorial-fetching-results.html
		// ... One common restriction is that SQLGetData may only be called on columns after the last bound column. ...
		if !binding {
			continue
		}
		bound, err := s.Cols[i].Bind(s.h, i)
		if err != nil {
			return err
		}
		if !bound {
			binding = false
		}
	}
	return nil
}
