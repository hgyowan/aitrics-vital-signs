package error

import (
	"net/http"
)

type Code int

const (
	None Code = 0

	Create Code = 100001
	Update Code = 100002
	Delete Code = 100003
	Upsert Code = 100004
	Get    Code = 100005

	WrongParam Code = 400001
	Conflict   Code = 400002
)

var businessCodeMap = map[Code]Status{
	None:       {int(None), http.StatusInternalServerError, "not exists error", nil, nil},
	Create:     {int(Create), http.StatusInternalServerError, "fail to create data", nil, nil},
	Update:     {int(Update), http.StatusInternalServerError, "fail to update data", nil, nil},
	Delete:     {int(Delete), http.StatusInternalServerError, "fail to delete data", nil, nil},
	Upsert:     {int(Upsert), http.StatusInternalServerError, "fail to upsert data", nil, nil},
	Get:        {int(Get), http.StatusInternalServerError, "fail to get data", nil, nil},
	WrongParam: {int(WrongParam), http.StatusBadRequest, "wrong parameter", nil, nil},
	Conflict:   {int(Conflict), http.StatusConflict, "conflict data", nil, nil},
}
