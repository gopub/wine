package wine_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/natande/wine"
)

var _testServer *wine.Server

type testJSONObj struct {
	Name string
	Age  int
}

func TestMain(m *testing.M) {
	_testServer = wine.DefaultServer()
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
	_testServer.Get("/json", func(c wine.Context) {
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

func TestHTML(t *testing.T) {
	var htmlText = `
	<html>
		<header>
		</header>
		<body>
			Hello, world!
		</body>
	</html>
	`
	_testServer.Get("/html/hello.html", func(c wine.Context) {
		c.HTML(htmlText)
	})

	resp, err := http.DefaultClient.Get("http://localhost:8000/html/hello.html")
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if string(data) != htmlText {
		t.Fatal(string(data))
	}

	if resp.Header["Content-Type"][0] != "text/html; charset=utf-8" {
		t.Fatal(resp.Header["Content-Type"])
	}
}

func TestPathParams(t *testing.T) {
	_testServer.Get("/sum/:a,:b", func(c wine.Context) {
		a := c.Params().Int("a")
		b := c.Params().Int("b")
		c.Text(fmt.Sprint(a + b))
	})

	_testServer.Get("/sum/:a,:b,:c", func(c wine.Context) {
		a := c.Params().Int("a")
		b := c.Params().Int("b")
		cc := c.Params().Int("c")
		c.Text(fmt.Sprint(a + b + cc))
	})

	{
		resp, err := http.DefaultClient.Get("http://localhost:8000/sum/1,2")
		if err != nil {
			t.Fatal(err)
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if string(data) != "3" {
			t.Fatal(string(data))
		}
	}

	{
		resp, err := http.DefaultClient.Get("http://localhost:8000/sum/1,2,-9")
		if err != nil {
			t.Fatal(err)
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if string(data) != "-6" {
			t.Fatal(string(data))
		}
	}

	{
		resp, err := http.DefaultClient.Get("http://localhost:8000/sum/1")
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Fatal(resp.Status)
		}
	}
}
