package wine

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

var _testServer *Server

type testJSONObj struct {
	Name string
	Age  int
}

func TestMain(m *testing.M) {
	_testServer = DefaultServer()
	go func() {
		_testServer.Run(":8000")
	}()
	result := m.Run()
	_testServer.Shutdown()
	os.Exit(result)
}

func TestJSON(t *testing.T) {
	obj := &testJSONObj{
		Name: "tom",
		Age:  19,
	}
	_testServer.Get("/json", func(c Context) {
		c.JSON(obj)
	})

	resp, err := http.DefaultClient.Get("http://localhost:8000/json")
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	var result testJSONObj
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatal(err)
	}

	if result != *obj {
		t.Fatal(result, *obj)
	}

	if resp.Header["Content-Type"][0] != "application/json; charset=utf-8" {
		t.Fatal(resp.Header["Content-Type"])
	}
}
