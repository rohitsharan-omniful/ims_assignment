package http

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omniful/go_commons/constants"
	"github.com/omniful/go_commons/env"
	"github.com/omniful/go_commons/log"
)

type LoggingMiddlewareOptions struct {
	Format      string
	Level       string
	LogRequest  bool
	LogResponse bool
	LogHeader   bool
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLogMiddleware(opts LoggingMiddlewareOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Ignore Health Requests
		if path == "/health" {
			c.Next()

			return
		}

		query := c.Request.URL.RawQuery
		start := time.Now()
		requestBodyString := "<Disabled>"
		bodyWriter := &responseWriter{
			body: bytes.NewBufferString("<Disabled>"),
		}

		// Create a custom ResponseWriter to capture the response body
		if opts.LogResponse {
			bodyWriter.body = bytes.NewBufferString("")
			bodyWriter.ResponseWriter = c.Writer
			c.Writer = bodyWriter
		}

		if opts.LogRequest {
			requestBody, _ := c.GetRawData()
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			requestBodyString = string(requestBody)
		}

		defer func() {
			// Capture the request complete timestamp
			end := time.Now()

			logFields := []log.Field{
				log.String("Amzn-Trace-ID", c.GetHeader("x-amzn-trace-id")),
				log.String("Method", c.Request.Method),
				log.String("Domain", c.Request.Host),
				log.String("Path", path),
				log.Int("Status", c.Writer.Status()),
				log.String("RequestBody", requestBodyString),
				log.String("Query", query),
				log.String("ResponseBody", bodyWriter.body.String()),
				log.String("IP", GetUserIPAddress(c)),
				log.String("User-Agent", c.Request.UserAgent()),
				log.Duration("Latency", time.Since(start)),
				log.String("RequestReceivedAt", start.Format(time.RFC3339)),
				log.String("RequestCompletedAt", end.Format(time.RFC3339)),
			}

			if correlationID := c.GetHeader(constants.HeaderXOmnifulCorrelationID); len(correlationID) > 0 {
				logFields = append(logFields, log.String(constants.HeaderXOmnifulCorrelationID, correlationID))
			}

			if clientService := c.GetHeader(constants.HeaderXClientService); len(clientService) > 0 {
				logFields = append(logFields, log.String(constants.HeaderXClientService, clientService))
			}

			if opts.LogHeader {
				for k, v := range c.Request.Header {
					if len(v) > 0 {
						logFields = append(logFields, log.String(k, v[0]))
					}
				}
			}

			log.WithContext(c, logFields...).Info(path)
		}()

		c.Next()
	}
}

// LoggerContextMiddleware sets a logger with the request ID in the Gin context.
func LoggerContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := env.GetRequestID(c)
		l := log.DefaultLogger().With(
			log.String(constants.HeaderXOmnifulRequestID, reqID),
		)

		ctx := log.ContextWithLogger(c, l).(*gin.Context)
		ctx.Next()
	}
}
