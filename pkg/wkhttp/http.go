package wkhttp

import (
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sendgrid/rest"
)

type WKHttp struct {
	r    *gin.Engine
	pool sync.Pool
}

func New() *WKHttp {
	l := &WKHttp{
		r:    gin.Default(),
		pool: sync.Pool{},
	}
	l.r.SetTrustedProxies(nil)
	l.pool.New = func() interface{} {
		return allocateContext()
	}
	return l
}

func NewWithLogger(loggerHandler HandlerFunc) *WKHttp {
	l := &WKHttp{
		r:    gin.New(),
		pool: sync.Pool{},
	}
	l.r.Use(l.LMHttpHandler(loggerHandler))
	l.r.Use(gin.Recovery())
	l.r.SetTrustedProxies(nil)
	l.pool.New = func() interface{} {
		return allocateContext()
	}
	return l
}

func (l *WKHttp) GetGinRoute() *gin.Engine {
	return l.r
}

func (l *WKHttp) Static(relativePath string, root string) {
	l.r.Static(relativePath, root)
}
func allocateContext() *Context {
	return &Context{Context: nil}
}

func (l *WKHttp) Use(handlers ...HandlerFunc) {
	l.r.Use(l.handlersToGinHandleFuncs(handlers)...)
}

func (l *WKHttp) handlersToGinHandleFuncs(handlers []HandlerFunc) []gin.HandlerFunc {
	newHandlers := make([]gin.HandlerFunc, 0, len(handlers))
	for _, handler := range handlers {
		newHandlers = append(newHandlers, l.LMHttpHandler(handler))
	}
	return newHandlers
}

type Context struct {
	*gin.Context
}

func (c *Context) reset() {
	c.Context = nil
}

func (c *Context) ResponseError(err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"msg":    err.Error(),
		"status": http.StatusBadRequest,
	})
}

func (c *Context) ResponseErrorWithStatus(status int, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"msg":    err.Error(),
		"status": status,
	})
}

func (c *Context) ResponseOK() {
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
	})
}

func (c *Context) ResponseOKWithData(data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"data":   data,
	})
}

func (c *Context) ResponseData(status int, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"data":   data,
	})
}

func (c *Context) ResponseStatus(status int) {
	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

func (c *Context) ForwardWithBody(url string, body []byte) {
	queryMap := map[string]string{}
	values := c.Request.URL.Query()
	for key, value := range values {
		queryMap[key] = value[0]
	}
	req := rest.Request{
		Method:      rest.Method(strings.ToUpper(c.Request.Method)),
		BaseURL:     url,
		Headers:     c.CopyRequestHeader(c.Request),
		Body:        body,
		QueryParams: queryMap,
	}

	resp, err := rest.API(req)
	if err != nil {
		c.ResponseError(err)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write([]byte(resp.Body))
}

func (c *Context) Forward(url string) {
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.ForwardWithBody(url, bodyBytes)
}

func (c *Context) CopyRequestHeader(request *http.Request) map[string]string {
	headerMap := map[string]string{}
	for key, values := range request.Header {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}
	return headerMap
}

func (c *Context) Username() string {
	return c.GetString("username")
}

type HandlerFunc func(c *Context)

func (l *WKHttp) LMHttpHandler(handlerFunc HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		hc := l.pool.Get().(*Context)
		hc.reset()
		hc.Context = c
		handlerFunc(hc)
		l.pool.Put(hc)
	}
}

func (l *WKHttp) Run(addr ...string) error {
	return l.r.Run(addr...)
}

func (l *WKHttp) POST(relativePath string, handlers ...HandlerFunc) {
	l.r.POST(relativePath, l.handlersToGinHandleFunc(handlers)...)
}

func (l *WKHttp) GET(relativePath string, handlers ...HandlerFunc) {
	l.r.GET(relativePath, l.handlersToGinHandleFunc(handlers)...)
}

func (l *WKHttp) DELETE(relativePath string, handlers ...HandlerFunc) {
	l.r.DELETE(relativePath, l.handlersToGinHandleFunc(handlers)...)
}

func (l *WKHttp) Any(relativePath string, handlers ...HandlerFunc) {
	l.r.Any(relativePath, l.handlersToGinHandleFunc(handlers)...)
}

func (l *WKHttp) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	l.r.ServeHTTP(w, req)
}

func (l *WKHttp) Group(relativePath string, handlers ...HandlerFunc) {
	l.r.Group(relativePath, l.handlersToGinHandleFunc(handlers)...)
}

func (l *WKHttp) handlersToGinHandleFunc(handlers []HandlerFunc) []gin.HandlerFunc {
	newHandlers := make([]gin.HandlerFunc, 0, len(handlers))
	for _, handler := range handlers {
		newHandlers = append(newHandlers, l.LMHttpHandler(handler))
	}
	return newHandlers
}

func CORSMiddleware() HandlerFunc {

	return func(c *Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length, Accept-Encoding, X-CSRF-Token, token, accept, origin, Cache-Control, X-Requested-With, appid, noncestr, sign, timestamp")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT,DELETE,PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
