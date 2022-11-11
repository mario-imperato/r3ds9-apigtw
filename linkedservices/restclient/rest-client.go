package restclient

import (
	"crypto/tls"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util"
	"github.com/go-resty/resty/v2"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"strconv"
)

type LinkedService struct {
	Cfg *Config
}

func NewInstanceWithConfig(cfg *Config) (*LinkedService, error) {
	lks := &LinkedService{Cfg: cfg}
	return lks, nil
}

func (lks LinkedService) NewClient(opts ...Option) (*Client, error) {
	cli := NewClient(lks.Cfg, opts...)
	return cli, nil
}

type Client struct {
	cfg Config

	restClient *resty.Client
	span       opentracing.Span
	spanOwned  bool
}

func NewClient(cfg *Config, opts ...Option) *Client {

	var clientOptions Config
	if cfg == nil {
		clientOptions = Config{TraceOpName: "rest-client"}
	} else {
		clientOptions = *cfg
	}

	for _, o := range opts {
		o(&clientOptions)
	}

	s := &Client{cfg: clientOptions}
	if clientOptions.Span != nil {
		s.cfg.NestTraceSpans = true
		s.span = clientOptions.Span
		s.spanOwned = false
	}

	s.restClient = resty.New()
	s.restClient.SetTimeout(s.cfg.RestTimeout)
	if s.cfg.SkipVerify {
		s.restClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return s
}

func (s *Client) Close() {
	if s.span != nil && s.spanOwned {
		s.span.Finish()
	}
}

func (s *Client) JsonRequest(reqDef *Request) (*Response, error) {

	reqSpan := s.getRequestSpan()
	defer reqSpan.Finish()

	reqDef.Headers = append(reqDef.Headers, NameValuePair{Name: "Accept", Value: "application/json"})
	req := s.getRequestWithSpan(reqDef, reqSpan)

	var resp *resty.Response
	var err error

	u := reqDef.URL
	switch reqDef.Method {
	case http.MethodGet:
		resp, err = req.Get(u)
	case http.MethodHead:
		resp, err = req.Head(u)
	case http.MethodPost:
		if reqDef.HasBody() {
			req = req.SetBody(reqDef.PostData.Data)
		}
		resp, err = req.Post(u)

	case http.MethodPut:
		if reqDef.HasBody() {
			req = req.SetBody(reqDef.PostData.Data)
		}
		resp, err = req.Put(u)

	}

	s.setSpanTags(reqSpan, u, reqDef.Method, resp.StatusCode(), err)
	if err != nil {
		err = util.NewError(strconv.Itoa(resp.StatusCode()), err)
		return NewResponse(http.StatusInternalServerError, "Error", "text/plain", []byte(err.Error()), nil), err
	}

	r := &Response{
		Status:      resp.StatusCode(),
		HTTPVersion: "1.1",
		StatusText:  resp.Status(),
		HeadersSize: -1,
		BodySize:    resp.Size(),
		Cookies:     []Cookie{},
		Content: &Content{
			MimeType: resp.Header().Get("Content-type"),
			Size:     resp.Size(),
			Data:     resp.Body(),
		},
	}

	for n, _ := range resp.Header() {
		r.Headers = append(r.Headers, NameValuePair{Name: n, Value: resp.Header().Get(n)})
	}

	return r, nil
	// return resp.StatusCode(), resp.Body(), resp.Header(), err
}

func (s *Client) getRequestWithSpan(reqDef *Request, reqSpan opentracing.Span) *resty.Request {

	req := s.restClient.R()
	// Transmit the span's TraceContext as HTTP headers on our outbound request.
	_ = opentracing.GlobalTracer().Inject(reqSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

	for _, h := range reqDef.Headers {
		req.SetHeader(h.Name, h.Value)
	}

	return req
}

func (s *Client) getRequestSpan() opentracing.Span {

	var reqSpan opentracing.Span

	if s.cfg.NestTraceSpans {
		if s.span == nil {
			s.span = opentracing.StartSpan(
				s.cfg.TraceOpName,
			)
			s.spanOwned = true
		}

		var parentCtx opentracing.SpanContext
		if s.span != nil {
			parentCtx = s.span.Context()
		}

		reqSpan = opentracing.StartSpan(
			s.cfg.TraceOpName,
			opentracing.ChildOf(parentCtx),
		)
	} else {
		reqSpan = opentracing.StartSpan(
			s.cfg.TraceOpName,
		)
	}

	return reqSpan
}

func (s *Client) setSpanTags(reqSpan opentracing.Span, endpoint, method string, statusCode int, err error) {

	reqSpan.SetTag(util.HttpUrlTraceTag, endpoint)
	reqSpan.SetTag(util.HttpMethodTraceTag, method)
	reqSpan.SetTag(util.HttStatusCodeTraceTag, statusCode)

	if err != nil {
		reqSpan.SetTag("error", err.Error())
	}
}
