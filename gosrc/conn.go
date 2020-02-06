package gosrc

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/alexbrainman/odbc/api"
)

// Conn struct to maintain database connection handle
type Conn struct {
	h api.SQLHDBC
	//tx  *Tx
	bad bool
}

// Connect the database using connection string
func (d *Driver) Connect(dsn string) (*Conn, error) {
	var out api.SQLHANDLE
	ret := api.SQLAllocHandle(api.SQL_HANDLE_DBC, api.SQLHANDLE(d.h), &out)
	h := api.SQLHDBC(out)
	err := Drv.Stats.updateHandleCount(api.SQL_HANDLE_DBC, 1)
	if err != nil {
		return nil, err
	}
	b := api.StringToUTF16(dsn)
	var outstrlen api.SQLSMALLINT
	var outst string
	outstr := api.StringToUTF16(outst)
	ret = api.SQLDriverConnect(h, 0,
		(*api.SQLWCHAR)(unsafe.Pointer(&b[0])), api.SQL_NTS,
		(*api.SQLWCHAR)(unsafe.Pointer(&outstr[0])), 1024, &outstrlen, api.SQL_DRIVER_NOPROMPT)
	if IsError(ret) {
		defer releaseHandle(h)
		fmt.Println("connection failed")
		return nil, NewError("SQLDriverConnect", h)
	}
	return &Conn{h: h}, nil
}

// Close close the connection handle database connection handle
func (c *Conn) Close() (err error) {
	h := c.h
	defer func() {
		c.h = api.SQLHDBC(api.SQL_NULL_HDBC)
		e := releaseHandle(h)
		if err == nil {
			err = e
		}
	}()
	ret := api.SQLDisconnect(c.h)
	if IsError(ret) {
		return c.newError("SQLDisconnect", h)
	}
	return err
}

func (c *Conn) newError(apiName string, handle interface{}) error {
	err := NewError(apiName, handle)
	if err != nil {
		log.Print(err)
	}
	return err
}
