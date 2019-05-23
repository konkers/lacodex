package lacodex

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/cskr/pubsub"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type testData struct {
	val int
	err error
}

func (td *testData) Get(w io.Writer) error {
	if td.err != nil {
		return td.err
	}
	fmt.Fprintf(w, "%d", td.val)
	td.val++
	return nil
}

func newTestWsConn(t *testing.T, serverURL string) *websocket.Conn {
	wsURL, _ := url.Parse(serverURL)
	wsURL.Scheme = "ws"
	wsURL.RawQuery = "async"

	c, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		t.Fatal("dial:", err)
	}
	return c
}

func testWsRead(t *testing.T, c *websocket.Conn) string {
	_, message, err := c.ReadMessage()
	if err != nil {
		t.Fatal("read:", err)
	}

	return string(message)
}

func testGet(t *testing.T, url string) string {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func testBadGet(t *testing.T, url string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Fatal("expected error from server")
	}

}

func TestWsHandler(t *testing.T) {
	ps := pubsub.New(0)
	td := &testData{}

	ts := httptest.NewServer(WsHandler(ps, td.Get))
	defer ts.Close()

	// Two successive requests should have incremeting responses
	r := testGet(t, ts.URL)
	assert.Equal(t, "0", r)

	r = testGet(t, ts.URL)
	assert.Equal(t, "1", r)

	// An async request w/o ws:// protocol should fail
	badURL, _ := url.Parse(ts.URL)
	badURL.RawQuery = "async"
	testBadGet(t, badURL.String())

	// Open a websocket connection
	c := newTestWsConn(t, ts.URL)
	defer c.Close()

	// Should get current state immediately.
	assert.Equal(t, "2", testWsRead(t, c))

	// Publishing an Update should give us the next state
	ps.Pub(nil, "update")
	assert.Equal(t, "3", testWsRead(t, c))

	ps.Pub(nil, "exit")
	_, _, err := c.ReadMessage()
	if err == nil {
		t.Fatal("expected read error after exit")
	}
}

func TestWsHandlerClientClose(t *testing.T) {
	ps := pubsub.New(0)
	td := &testData{}

	ts := httptest.NewServer(WsHandler(ps, td.Get))
	defer ts.Close()

	// Open a websocket connection
	c := newTestWsConn(t, ts.URL)

	assert.Equal(t, "0", testWsRead(t, c))
	// Close the connection early
	c.Close()

	// There's not a great way to ensure proper cleanup here.
}

func TestWsHandlerSyncGetError(t *testing.T) {
	ps := pubsub.New(0)
	td := &testData{
		err: fmt.Errorf("Forced error."),
	}

	ts := httptest.NewServer(WsHandler(ps, td.Get))
	defer ts.Close()

	testBadGet(t, ts.URL)
}

func TestWsHandlerSyncAsyncGetFirstError(t *testing.T) {
	ps := pubsub.New(0)
	td := &testData{
		err: fmt.Errorf("Forced error."),
	}

	ts := httptest.NewServer(WsHandler(ps, td.Get))
	defer ts.Close()

	// Open a websocket connection
	c := newTestWsConn(t, ts.URL)
	defer c.Close()

	_, _, err := c.ReadMessage()
	if err == nil {
		t.Fatal("Expected Error")
	}
}

func TestWsHandlerSyncAsyncGetSecondError(t *testing.T) {
	ps := pubsub.New(0)
	td := &testData{}

	ts := httptest.NewServer(WsHandler(ps, td.Get))
	defer ts.Close()

	// Open a websocket connection
	c := newTestWsConn(t, ts.URL)
	defer c.Close()

	// Initial read should work.
	assert.Equal(t, "0", testWsRead(t, c))

	// Now set the error and try another read.
	td.err = fmt.Errorf("Forced error.")
	ps.Pub(nil, "update")
	_, _, err := c.ReadMessage()
	if err == nil {
		t.Fatal("Expected Error")
	}
}

func TestWsHandlerPing(t *testing.T) {
	// Set timeouts to be super agressive so we can fail fast.
	wsSetTimeouts(time.Millisecond*10, time.Millisecond*100)
	defer wsSetTimeouts(10*time.Second, 60*time.Second)

	ps := pubsub.New(0)
	td := &testData{}

	ts := httptest.NewServer(WsHandler(ps, td.Get))
	defer ts.Close()

	// Open a websocket connection
	c := newTestWsConn(t, ts.URL)
	defer c.Close()

	//  Respond to the first ping.
	pingCount := 0
	c.SetPingHandler(func(s string) error {
		if pingCount == 0 {
			c.WriteMessage(websocket.PongMessage, []byte(s))
		}
		pingCount++
		return nil
	})

	// Initial read should work.
	assert.Equal(t, "0", testWsRead(t, c))

	// Second read should fail.
	_, _, err := c.ReadMessage()
	if err == nil {
		t.Fatal("Expected error.")
	}
}
