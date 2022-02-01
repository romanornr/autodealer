package webserver

//
//import (
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/go-chi/chi/v5"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestServer(t *testing.T) {
//	w := httptest.NewRecorder()
//	r, err := http.NewRequest("GET", "/test", nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	router := chi.NewRouter()
//	router.Get("/hi", func(w http.ResponseWriter, r *http.Request) {
//		w.Write([]byte("hi"))
//	})
//
//	// Serve the HTTP request.
//	router.ServeHTTP(w, r)
//	assert.Equal(t, "hi", http.StatusOK)
//}
