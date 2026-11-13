package biz

import "context"

type TokenResult struct {
	AccessToken string `json:"accessToken"`
}

type PingReq struct {
	Name string `json:"name"`
}

type PingResp struct {
	Ok bool `json:"ok"`
}

type GetBookRequest struct {
	Id int64 `json:"id"`
}

type GetBookResponse struct {
	Name string `json:"name"`
}

type GetOrderRequest struct {
	Id int64 `json:"id"`
}

type GetOrderResponse struct {
	Id int64 `json:"id"`
}

type CreateOrderRequest struct {
	Name string `json:"name"`
}

type CreateOrderResponse struct {
	Id int64 `json:"id"`
}

// GetTokenHost returns the prpc token service base URL.
// It is called once when TokenClient is constructed.
func GetTokenHost() string {
	return "http://token.internal"
}

// GetBookHost returns the srpc book service base URL.
// It is called once when BookClient is constructed.
func GetBookHost() string {
	return "http://book.internal"
}

// GetOrderHost returns the restrpc order service base URL.
// It is called when OrderSvc is constructed.
func GetOrderHost() string {
	return "http://order.internal"
}

// InitHeader builds the headers sent on every TokenClient request.
// The prpc client calls it before each request.
func InitHeader(ctx context.Context) map[string]string {
	return map[string]string{"X-App": "gos-example"}
}

// TokenClient is a prpc client. header names the function called before each request.
// @gos type=prpc; host=GetTokenHost(); header=InitHeader
type TokenClient interface {
	// @gos url="/token/get"
	GetToken(ctx context.Context, tokenStr string) (*TokenResult, error)
}

// BookClient is an srpc client. Responses use the servlet envelope {code,msg,obj}.
// @gos type=srpc; host=GetBookHost()
type BookClient interface {
	// @gos url="/book/get"
	GetBook(ctx context.Context, req *GetBookRequest) (*GetBookResponse, error)

	// @gos url="/book/list"
	ListBook(ctx context.Context, req *GetBookRequest) (*GetBookResponse, error)
}

// OrderClient is a restrpc client. HTTP status decides success, and GET encodes the request as a query.
// @gos type=restrpc; host=GetOrderHost()
type OrderClient interface {
	// @gos url="/order/get"; method=GET
	GetOrder(ctx context.Context, req *GetOrderRequest) (*GetOrderResponse, error)

	// @gos url="/order/create"
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
}

// NoteInner is the prpc server. Its group defaults to prpc, matching the prpc filters.
// @gos type=prpc; url="/prpc"; group=prpc
type NoteInner struct{}

// Ping answers an internal prpc call.
// It reports Ok when the request has a name.
// @gos url="/ping"
func (n *NoteInner) Ping(ctx context.Context, req *PingReq) (*PingResp, error) {
	return &PingResp{Ok: req != nil && req.Name != ""}, nil
}

// loadBook calls the default book host, or another host when host is set.
// BookSvc is filled because it is marked auto; BookSvcWithPrefix returns a new client.
func loadBook(ctx context.Context, host string, id int64) (*GetBookResponse, error) {
	req := &GetBookRequest{Id: id}
	if host != "" {
		return BookSvcWithPrefix(host).GetBook(ctx, req)
	}
	return BookSvc.GetBook(ctx, req)
}

// TokenSvc is filled from the generated prpc client because it is marked auto.
// @gos auto
var TokenSvc TokenClient

// BookSvc is the default srpc client. BookSvcWithPrefix is assigned for other hosts.
// @gos auto
var BookSvc BookClient

var BookSvcWithPrefix func(prefix string) BookClient

// OrderSvc is assigned by the restrpc initializer. restrpc globals do not need auto.
var OrderSvc OrderClient
