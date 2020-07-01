package casinoapi

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gode/types"
)

func TestFlash2db_Call(t *testing.T) {
	t.Run("get correct url and return", func(t *testing.T) {
		service := "casino.slot.line243.BuBuGaoSheng"
		function := "beginGame"
		sid := types.SessionID(`19870604xi`)
		uid := types.UserID(9527)
		betInfo := types.BetInfo(`{"BetLevel":5}`)
		credit := types.Credit(50000)
		APIResult := `{"event": true}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// "/amfphp/json.php/casino.slot.line243.BuBuGaoSheng.beginGame/19870604xi/9527/{"BetLevel":5}/50000"
			want := fmt.Sprintf("%s/%s.%s/%s/%d/%s/%d", urlPrefix, service, function, sid, uid, betInfo, credit)
			assertURLEqual(t, r.URL.Path, want)

			_, _ = fmt.Fprint(w, APIResult)
		}))

		f := NewFlash2db(server.URL)
		gotResult, _ := f.Call(service, function, sid, uid, betInfo, credit)

		if bytes.Compare([]byte(APIResult), gotResult) != 0 {
			t.Errorf("want %s, got %s", APIResult, gotResult)
		}
	})

	t.Run("returns error when connect failed", func(t *testing.T) {
		f := NewFlash2db("http://not.exists")
		_, err := f.Call("dummyServer", "dummyFunc", "dummyParam")

		if err == nil {
			t.Errorf("expected an error but not got one")
		}
	})

	t.Run("returns error when not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`1231345fg`))
		}))

		f := NewFlash2db(server.URL)
		_, err := f.Call("dummyServer", "dummyFunc", "dummyParam")

		if err == nil {
			t.Errorf("expected an error but not got one")
		}
	})

	t.Run("returns error when got 500 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}))

		f := NewFlash2db(server.URL)
		_, err := f.Call("dummyServer", "dummyFunc", "dummyParam")

		if err == nil {
			t.Errorf("expected an error but not got one")
		}
	})
}

func TestMakeURL(t *testing.T) {
	f := &Flash2db{}

	got := f.makeURL("Client", "CheckLogin", "someSID", "127.0.0.1")
	want := "/amfphp/json.php/Client.CheckLogin/someSID/127.0.0.1"

	assertURLEqual(t, got, want)
}

func assertURLEqual(t *testing.T, got string, want string) {
	t.Helper()
	if got != want {
		t.Errorf("URL not equal,\nwant %q\n got %q", want, got)
	}
}
