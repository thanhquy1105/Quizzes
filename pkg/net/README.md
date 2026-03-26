

useage

```go


 net.Serve("tcp://127.0.0.1:1999",handler,opts...)

```

event

```go
OnConnect(c *net.Conn)

OnClose(c *net.Conn)

OnData(c *net.Conn)

```