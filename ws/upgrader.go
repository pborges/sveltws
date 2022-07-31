package ws

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Upgrader struct {
	Log          *log.Logger
	OnConnect    func(conn net.Conn)
	OnDisconnect func(conn net.Conn)
	OnMessage    func(ctx *Context, buf []byte)
}

func (u *Upgrader) logln(v ...any) {
	if u.Log != nil {
		u.Log.Output(2, strings.TrimSpace(fmt.Sprintln(v...)))
	}
}

func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		u.logln("[WS !!]", err.Error())
		return
	}
	u.logln("[WS !!] Connect", conn.RemoteAddr().String())
	if u.OnConnect != nil {
		u.OnConnect(conn)
	}

	ctx := Context{
		conn: conn,
		u:    u,
	}

	defer func() {
		u.logln("[WS !!] Disconnect", conn.RemoteAddr().String())
		if u.OnDisconnect != nil {
			u.OnDisconnect(conn)
		}
		ctx.Close()
	}()

	for {
		msg, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			return
		}
		u.logln("[WS >>]", string(msg))
		if u.OnMessage != nil {
			u.OnMessage(&ctx, msg)
		}
	}
}

type Context struct {
	conn    net.Conn
	u       *Upgrader
	cleanup []func()
	mu      sync.Mutex
}

func (c *Context) OnDisconnect(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cleanup = append(c.cleanup, fn)
}

func (c *Context) Write(buf []byte) (int, error) {
	c.u.logln("[WS <<]", string(buf))
	return len(buf), wsutil.WriteServerMessage(c.conn, ws.OpText, buf)
}

func (c *Context) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.u.logln("[WS !!] Close", "cleanup:", len(c.cleanup))
	for _, fn := range c.cleanup {
		fn()
	}
	return c.conn.Close()
}

func (c *Context) CloseWithError(err error) {
	c.u.logln("[WS !!] Close with error", err.Error())
	c.Close()
}
