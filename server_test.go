package wine_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gopub/wine"
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
	server.Get("/json", func(c *wine.Context) {
		c.JSON(http.StatusOK, obj)
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
	server.Get("/html/hello.html", func(c *wine.Context) {
		c.HTML(http.StatusOK, htmlText)
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
	server.Get("/sum/:a,:b", func(c *wine.Context) {
		a := c.Params().Int("a")
		b := c.Params().Int("b")
		c.Text(http.StatusOK, fmt.Sprint(a+b))
	})

	server.Get("/sum/:a,:b,:c", func(c *wine.Context) {
		a := c.Params().Int("a")
		b := c.Params().Int("b")
		cc := c.Params().Int("c")
		c.Text(http.StatusOK, fmt.Sprint(a+b+cc))
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
