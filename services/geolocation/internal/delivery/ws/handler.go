package ws

import (
	"time"

	"github.com/labstack/echo/v4"
)

// HandleTracking — GET /ws/tracking — WebSocket stub: accept, send "connected", read loop
func HandleTracking(hub *Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, err := hub.HandleConnect(c.Response(), c.Request())
		if err != nil {
			return err
		}
		defer hub.Unregister(conn)
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})
		if err := conn.WriteMessage(1, []byte(`{"type":"connected","service":"geolocation","message":"tracking stub"}`)); err != nil {
			return nil
		}
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
			// Stub: echo or ignore; later: parse driver location updates, broadcast to ride subscribers
		}
		return nil
	}
}
