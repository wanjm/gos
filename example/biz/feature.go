package biz

import (
	"context"
	"errors"
)

type PriceReq struct {
	PriceUserNum int    `json:"priceUserNum" default:"6"`
	OpenComment  bool   `json:"openComment" default:"true"`
	Title        string `json:"title" default:"draft"`
}

type AdminReq struct {
	PriceReq
	Name  string `json:"name"`
	AppID string `json:"appId" header:"X-AppId"`
}

type GetReq struct {
	Id string `json:"id"`
}

// Demo is a servlet. folder is the Apifox directory unless a method sets its own.
// Clock, Token, and Book are injected by type.
// @gos type=servlet; url="/demo"; group=servlet; folder="example/demo"
type Demo struct {
	Clock *Clock
	Token TokenClient
	Book  BookClient
}

// Create stores one note after the permission check and the AdminGate filter.
// It calls the injected clients and the restrpc OrderSvc variable.
// @gos url="/admin"; title="create note"; method=POST; folder="example/admin"; header="X-AppId:应用ID"; filters=AdminGate; permission=perm.PermGroupUser,perm.PermActionCreate
func (d *Demo) Create(ctx context.Context, req *AdminReq) (*NoteInfo, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}
	_, _ = d.Token.GetToken(ctx, req.AppID)
	_, _ = d.Book.GetBook(ctx, &GetBookRequest{Id: 1})
	_, _ = OrderSvc.CreateOrder(ctx, &CreateOrderRequest{Name: req.Name})
	_, _ = loadBook(ctx, "http://book-dev", 1)
	_ = noteFilters(req.Name, 1)
	_ = d.Clock
	return &NoteInfo{Name: req.Name, Title: req.Title}, nil
}

// Get reads one note. The :id path segment is copied into GetReq.Id.
// It returns a note whose name is that id.
// @gos url="/item/:id"; method=GET; title="get note"
func (d *Demo) Get(ctx context.Context, req *GetReq) (*NoteInfo, error) {
	return &NoteInfo{Name: req.Id}, nil
}

// Hidden is omitted from OpenAPI because the method sets swagger=false.
// It returns a fixed string.
// @gos url="/hidden"; title="hidden demo"; swagger=false
func (d *Demo) Hidden(ctx context.Context, req *GetReq) (string, error) {
	return "hidden", nil
}

// Internal is left out of OpenAPI because the struct sets swagger=false.
// @gos type=servlet; url="/internal"; swagger=false
type Internal struct{}

// Health is a servlet route that stays out of OpenAPI with its struct.
// It returns ok.
// @gos url="/health"; title="health"
func (i *Internal) Health(ctx context.Context, req *GetReq) (string, error) {
	return "ok", nil
}
