package live

import (
	"errors"
	"github.com/valyala/fasthttp"
	"regexp"
	"time"
)

type FastHttpHandler interface {
	FastDo() (*Room, error)
}

var (
	_ FastHttpHandler = (*fastHttpHandler)(nil)
)

type fastHttpHandler struct {
	client     *fasthttp.Client
	headers    map[string]string
	requestURI string
}

type WithOption func(client *fastHttpHandler)

func WithReadTimeout(readTimeout time.Duration) WithOption {
	return func(client *fastHttpHandler) {
		client.client.ReadTimeout = readTimeout
	}
}

func WithWriteTimeout(writeTimeout time.Duration) WithOption {
	return func(client *fastHttpHandler) {
		client.client.WriteTimeout = writeTimeout
	}
}
func WithMaxIdleConnDuration(maxIdleConnDuration time.Duration) WithOption {
	return func(client *fastHttpHandler) {
		client.client.MaxIdleConnDuration = maxIdleConnDuration
	}
}
func WithConcurrencyAndDNSCacheDuration(concurrency int, dcd time.Duration) WithOption {
	return func(client *fastHttpHandler) {
		client.client.Dial = (&fasthttp.TCPDialer{
			Concurrency:      concurrency,
			DNSCacheDuration: dcd,
		}).Dial
	}
}
func WithNoDefaultUserAgentHeader(agentHeader bool) WithOption {
	return func(client *fastHttpHandler) {
		client.client.NoDefaultUserAgentHeader = agentHeader
	}
}

func WithDisableHeaderNamesNormalizing(headerNames bool) WithOption {
	return func(client *fastHttpHandler) {
		client.client.DisableHeaderNamesNormalizing = headerNames
	}
}

func WithHeader(headers map[string]string) WithOption {
	return func(client *fastHttpHandler) {
		client.headers = headers
	}
}

func WithRequestURI(requestURI string) WithOption {
	return func(client *fastHttpHandler) {
		client.requestURI = requestURI
	}
}
func NewFastHttpHandler(opt ...WithOption) FastHttpHandler {
	handler := &fastHttpHandler{
		client: &fasthttp.Client{
			// 读超时时间, 不设置read超时,可能会造成连接复用失效
			ReadTimeout: time.Second * 5,
			// 写超时时间
			WriteTimeout: time.Second * 5,
			// 5秒后，关闭空闲的活动连接
			MaxIdleConnDuration: time.Second * 5,
			// 当true时,从请求中去掉User-Agent标头
			NoDefaultUserAgentHeader: true,
			// 当true时，header中的key按照原样传输，默认会根据标准化转化
			DisableHeaderNamesNormalizing: true,
			//当true时,路径按原样传输，默认会根据标准化转化
			DisablePathNormalizing: true,
			Dial: (&fasthttp.TCPDialer{
				// 最大并发数，0表示无限制
				Concurrency: 4096,
				//将 DNS 缓存时间从默认分钟增加到一小时
				DNSCacheDuration: time.Hour,
			}).Dial,
			ReadBufferSize: 4096 * 4096,
		},
		headers:    make(map[string]string),
		requestURI: "https://live.douyin.com/7003418886",
	}
	for _, o := range opt {
		o(handler)
	}
	return handler
}

func (h *fastHttpHandler) FastDo() (*Room, error) {
	req, resp, cookie := fasthttp.AcquireRequest(), fasthttp.AcquireResponse(), fasthttp.AcquireCookie()
	// 回收 实例到请求池
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseCookie(cookie)
	}()
	resp.Header.SetBytesV(fasthttp.HeaderContentType, []byte("application/json"))
	// 获取客户端连接
	req.Header.SetMethod(fasthttp.MethodGet)
	if len(h.headers) > 0 {
		for k, v := range h.headers {
			req.Header.Set(k, v)
		}
	}
	var (
		err error
	)
	req.SetRequestURI(h.requestURI)
	resp.Header.SetStatusCode(fasthttp.StatusOK)
	if err = h.client.Do(req, resp); err != nil {
		return nil, err
	}
	// 获取cookie
	cookie.SetKey("ttwid")
	resp.Header.Cookie(cookie)
	re := regexp.MustCompile(`roomId\\":\\"(\d+)\\"`)
	match := re.FindStringSubmatch(string(resp.Body()))
	if match == nil || len(match) < 2 {
		return nil, errors.New("No match found")
	}
	return &Room{
		RoomID: match[1],
		TtwID:  string(cookie.Value()),
		URL:    h.requestURI,
	}, nil
}
