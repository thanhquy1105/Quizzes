package wkhttp

import (
	"fmt"
	"time"

	"btaskee-quiz/pkg/wklog"
	"go.uber.org/zap"
)

func LoggerWithWklog(log wklog.Log) HandlerFunc {
	return func(c *Context) {

		start := time.Now()

		c.Next()

		latency := time.Since(start)

		if latency > time.Minute {
			latency = latency.Truncate(time.Second)
		}

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		log.Debug(fmt.Sprintf("|%s| %d| %s", c.Request.Method, c.Writer.Status(), path),
			zap.String("clientip", c.ClientIP()),
			zap.Int("size", c.Writer.Size()),
			zap.String("latency", fmt.Sprintf("%v", latency)))
	}
}
