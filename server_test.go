package wine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gopub/wine/v2"
)

var server *wine.Server

type testJSONObj struct {
	Name string
	Age  int
}

func TestMain(m *testing.M) {
	server = wine.DefaultServer()
	go func() {
		server.Run(":8000")
	}()

	time.Sleep(time.Second)
	result := m.Run()
	os.Exit(result)
}

func TestJSON(t *testing.T) {
	obj := &testJSONObj{
		Name: "tom",
		Age:  19,
	}
	server.Get("/json", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		responder.JSON(http.StatusOK, obj)
		return true
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
	server.Get("/html/hello.html", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		responder.HTML(http.StatusOK, htmlText)
		return true
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
	server.Get("/sum/:a,:b", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		a := request.Parameters().Int("a")
		b := request.Parameters().Int("b")
		responder.Text(http.StatusOK, fmt.Sprint(a+b))
		return true
	})

	server.Get("/sum/:a,:b,:c", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		a := request.Parameters().Int("a")
		b := request.Parameters().Int("b")
		cc := request.Parameters().Int("c")
		responder.Text(http.StatusOK, fmt.Sprint(a+b+cc))
		return true
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
