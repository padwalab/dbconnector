package gosrc

import (
	"github.com/alexbrainman/odbc/api"
)

var Drv Driver

// Driver struct to manange env handle
type Driver struct {
	Stats
	h api.SQLHENV
}

// Load the index.html template.
func initDriver() error {
	var out api.SQLHANDLE
	in := api.SQLHANDLE(api.SQL_NULL_HANDLE)
	ret := api.SQLAllocHandle(api.SQL_HANDLE_ENV, in, &out)
	if IsError(ret) {
		return NewError("SQLAllocHandle", api.SQLHENV(in))
	}
	Drv.h = api.SQLHENV(out)
	err := Drv.Stats.updateHandleCount(api.SQL_HANDLE_ENV, 1)
	if err != nil {
		return err
	}
	// will use ODBC v3
	ret = api.SQLSetEnvUIntPtrAttr(Drv.h, api.SQL_ATTR_ODBC_VERSION, api.SQL_OV_ODBC3, 0)
	if IsError(ret) {
		defer releaseHandle(Drv.h)
		return NewError("SQLSetEnvUIntPtrAttr", Drv.h)
	}
	return nil
}

// Close release the env handle
func (d *Driver) Close() error {
	// TODO(brainman): who will call (*Driver).Close (to dispose all opened handles)?
	h := d.h
	d.h = api.SQLHENV(api.SQL_NULL_HENV)
	return releaseHandle(h)
}

func init() {
	err := initDriver()
	if err != nil {
		panic(err)
	}
}
