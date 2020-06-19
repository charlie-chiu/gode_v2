package casinoapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFlash2db_Call(t *testing.T) {
	t.Run("call correct url", func(t *testing.T) {
		service := "casino.slot.line243.BuBuGaoSheng"
		function := "beginGame"
		param1 := "19870604xi"
		param2 := 7788

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			want := "/amfphp/json.php/casino.slot.line243.BuBuGaoSheng.beginGame/19870604xi/7788"

			assertURLEqual(t, r.URL.Path, want)
		}))

		host := server.URL

		f := NewFlash2db(host)
		_, _ = f.Call(service, function, param1, param2)
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
