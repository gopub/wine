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
        
## Path Parameters
Single parameter in one segment
<pre>
    s := wine.Default() <br/>
    s.Get("/items/<b>:id</b>", func(c wine.Context) { <br/>
        id := c.Params().GetStr("id") <br/>
        c.Text("item id: " + id) <br/>
    }) <br/>
    s.Run(":8000")
</pre>
        
Multiple parameters in one segment   
<pre>
    s := wine.Default() <br/>
    s.Get("/items/<b>:page,:size</b>", func(c wine.Context) { <br/>
        page := c.Params().GetInt("page") <br/>
        size := c.Params().GetInt("size") <br/>
        c.Text("page:" + strconv.Itoa(page) + " size:" + strconv.Itoa(size)) <br/>
    }) <br/>
    s.Run(":8000")
</pre>

## Use Middlewares
Use middlewares to intercept and preprocess requests  

    //Custom middleware
    func Logger(c wine.Context) {
    	st := time.Now()
    	
    	//call Next() to pass request to the next handler
    	c.Next() 
    	
    	cost := float32((time.Since(st) / time.Microsecond)) / 1000.0
    	req := c.HTTPRequest()
    	log.Printf("[WINE] %.3fms %s %s", cost, req.Method, req.RequestURI)
    }
    
    func main() {
    	s := wine.NewServer()
    	
    	//Use middleware Logger
    	s.Use(Logger)
    	
    	s.Get("/hello", func(c wine.Context) {
    		c.Text("Hello, Wine!")
        })
        s.Run(":8000")
    }
  