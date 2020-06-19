package casinoapi

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFlash2db_Call(t *testing.T) {
	t.Run("get correct url and return", func(t *testing.T) {
		service := "casino.slot.line243.BuBuGaoSheng"
		function := "beginGame"
		param1 := "19870604xi"
		param2 := "7788"
		APIResult := `{"event": true}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// "/amfphp/json.php/casino.slot.line243.BuBuGaoSheng.beginGame/19870604xi/7788"
			want := fmt.Sprintf("%s/%s.%s/%s/%s", urlPrefix, service, function, param1, param2)
			assertURLEqual(t, r.URL.Path, want)

			_, _ = fmt.Fprint(w, APIResult)
		}))

		f := NewFlash2db(server.URL)
		gotResult, _ := f.Call(service, function, param1, param2)

		if bytes.Compare([]byte(APIResult), gotResult) != 0 {
			t.Errorf("want %s, got %s", APIResult, gotResult)
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
