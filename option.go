package restgo

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type option struct {
	baseURL       *url.URL
	globalHeader  http.Header
	transport     http.RoundTripper
	jar           http.CookieJar
	timeout       time.Duration
	checkRedirect func(req *http.Request, via []*http.Request) error
	beforeHooks   []BeforeHookFunc
	afterHooks    []AfterHookFunc
}

type OptionFn func(opt *option)

func WithBaseURL(baseURL string) OptionFn {
	return func(opt *option) {
		opt.baseURL, _ = url.ParseRequestURI(baseURL)
	}
}

func WithGlobalHeader(header http.Header) OptionFn {
	return func(opt *option) {
		opt.globalHeader = header
	}
}

func WithTransport(transport http.RoundTripper) OptionFn {
	return func(opt *option) {
		opt.transport = transport
	}
}

func WithCert(certPool *x509.CertPool, cert tls.Certificate) OptionFn {
	return func(opt *option) {
		opt.transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      certPool,
				Certificates: []tls.Certificate{cert},
			},
		}
	}
}

func WithJar(jar *cookiejar.Jar) OptionFn {
	return func(opt *option) {
		opt.jar = jar
	}
}

func WithCookies(u *url.URL, cookies ...*http.Cookie) OptionFn {
	return func(opt *option) {
		if opt.jar == nil {
			opt.jar, _ = cookiejar.New(nil)
		}
		opt.jar.SetCookies(u, cookies)
	}
}

func WithTimeout(timeout time.Duration) OptionFn {
	return func(opt *option) {
		opt.timeout = timeout
	}
}

func WithCheckRedirect(checkRedirect func(req *http.Request, via []*http.Request) error) OptionFn {
	return func(opt *option) {
		opt.checkRedirect = checkRedirect
	}
}

// WithBeforeHook 挂载请求前的钩子函数
func WithBeforeHook(hook BeforeHookFunc) OptionFn {
	return func(opt *option) {
		opt.beforeHooks = append(opt.beforeHooks, hook)
	}
}

// WithAfterHook 挂载请求后的钩子函数
func WithAfterHook(hook AfterHookFunc) OptionFn {
	return func(opt *option) {
		opt.afterHooks = append(opt.afterHooks, hook)
	}
}
