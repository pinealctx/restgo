package restgo

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

const (
	version           = "0.0.1"
	defaultUA         = "RestGO/" + version
	headerContentType = "Content-Type"
	headerUserAgent   = "User-Agent"
)

type Client struct {
	baseURL      *url.URL
	globalHeader http.Header
	beforeHooks  []BeforeHookFunc
	afterHooks   []AfterHookFunc

	client *http.Client
}

func New(optFns ...OptionFn) *Client {
	var o = &option{}
	for _, fn := range optFns {
		fn(o)
	}
	return &Client{
		baseURL:      o.baseURL,
		globalHeader: o.globalHeader,
		beforeHooks:  o.beforeHooks,
		afterHooks:   o.afterHooks,
		client: &http.Client{
			Transport:     o.transport,
			Jar:           o.jar,
			Timeout:       o.timeout,
			CheckRedirect: o.checkRedirect,
		},
	}
}

func (c *Client) Do(ctx context.Context, req IRequest) (IResponse, error) {
	// run before hooks
	c.runBeforeHooks(req)

	var rURL, err = req.MakeURL(CloneURL(c.baseURL))
	if err != nil {
		return nil, err
	}
	var body io.Reader
	body, err = req.MakeRequestBody()
	if err != nil {
		return nil, err
	}
	var request *http.Request
	request, err = http.NewRequestWithContext(ctx, req.GetMethod(), rURL, body)
	if err != nil {
		return nil, err
	}
	for k, vList := range c.globalHeader {
		for _, v := range vList {
			request.Header.Set(k, v)
		}
	}
	req.WrapperHTTPRequest(request)
	var ua = request.Header.Get(headerUserAgent)
	if ua == "" {
		request.Header.Set(headerUserAgent, defaultUA)
	}
	var response *http.Response
	response, err = c.client.Do(request)
	if err != nil {
		return nil, err
	}
	var rsp = NewResponse(response)
	// run after hooks
	c.runAfterHooks(req, rsp)
	return rsp, nil
}

func (c *Client) Execute(ctx context.Context, method, resource string, params IParams) (IResponse, error) {
	var req = NewRequest(method, resource)
	req.AddParams(params.Params()...)
	return c.Do(ctx, req)
}

func (c *Client) Get(ctx context.Context, resource string, params IParams) (IResponse, error) {
	return c.Execute(ctx, "GET", resource, params)
}

func (c *Client) Post(ctx context.Context, resource string, params IParams) (IResponse, error) {
	return c.Execute(ctx, "POST", resource, params)
}

func (c *Client) Put(ctx context.Context, resource string, params IParams) (IResponse, error) {
	return c.Execute(ctx, "PUT", resource, params)
}

func (c *Client) runBeforeHooks(req IRequest) {
	for _, hook := range c.beforeHooks {
		hook(req)
	}
}

func (c *Client) runAfterHooks(req IRequest, rsp IResponse) {
	for _, hook := range c.afterHooks {
		hook(req, rsp)
	}
}
