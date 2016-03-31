# WINE

Wine is a lightweight web framework for quickly writing web applications/services in Golang. 

## Install  

        $ go get -u -v github.com/justintan/wine

## Quick start  
Create ./hello.go  
        
        package main
        
        import "github.com/justintan/wine"
        
        func main() {
        	s := wine.Default()
        	s.Get("/hello", func(c wine.Context) {
        		c.Text("Hello, Wine!")
        	})
        	s.Run(":8000")
        }
Run and test:  

        $ go run hello.go
        $ curl http://localhost:8000/hello
        $ Hello, Wine!
        

## JSON Rendering

        s := wine.Default()
        s.Get("/time", func(c wine.Context) {
        	c.JSON(map[string]interface{}{"time":time.Now().Unix()})
        })
        s.Run(":8000")

## Parameters
Context.Params() provides an uniform interface to retrieve request parameters, which might be in query string, http body, url path, etc. Form and json are supported.  

        s := wine.Default()
        s.Post("feedback", func(c wine.Context) {
            text := c.Params().GetStr("text")
            email := c.Params().GetStr("email")
            c.Text("Feedback:" + text + " from " + email)
        })
        s.Run(":8000")
Test parameters in query string

        $ curl -X POST "http://localhost:8000/feedback?text=crash&email=wine@wine.com"
Test parameters in form

        $ curl -X POST -d "text=crash&email=wine@wine.com" http://localhost:8000/feedback
Test parameters in json

        $ curl -X POST -H "Content-Type:application/json" -d '{"text":"crash", "email":"wine@wine.com"}' http://localhost:8000/feedback
#### Parameters in URL Path
Path parameters are also supported in order to provide elegant RESTFul api.
Single parameter in one segment:
<pre>
    s := wine.Default() 
    s.Get("/items/<b>:id</b>", func(c wine.Context) { 
        id := c.Params().GetStr("id") 
        c.Text("item id: " + id) 
    }) 
    s.Run(":8000")
</pre>
        
Multiple parameters in one segment:   
<pre>
    s := wine.Default() 
    s.Get("/items/<b>:page,:size</b>", func(c wine.Context) { 
        page := c.Params().GetInt("page") 
        size := c.Params().GetInt("size") 
        c.Text("page:" + strconv.Itoa(page) + " size:" + strconv.Itoa(size)) 
    }) 
    s.Run(":8000")
</pre>

## Use Middlewares
Use middlewares to intercept and preprocess requests  

Custom middleware
<pre>
    func Logger(c wine.Context) {
    	st := time.Now()  
    	//pass request to the next handler
    	<b>c.Next()</b> 
    	cost := float32((time.Since(st) / time.Microsecond)) / 1000.0
    	req := c.HTTPRequest()
    	log.Printf("[WINE] %.3fms %s %s", cost, req.Method, req.RequestURI)
    } <br/>
    func main() {
    	s := wine.NewServer() 
    	//Use middleware Logger
    	<b>s.Use(Logger)</b> 
    	s.Get("/hello", func(c wine.Context) {
    		c.Text("Hello, Wine!")
        })
        s.Run(":8000")
    }
</pre>
  