# WINE

Wine is a lightweight web framework for quickly writing web applications/services in Go. 

## Install  

        $ go get -u -v github.com/gopub/wine

## Quick start  
Create ./hello.go  
        
        package main
        
        import "github.com/gopub/wine"
        
        func main() {
        	s := wine.NewServer()
        	s.Get("/hello", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder {
        		return wine.Text("Hello, Wine!")
        	})
        	s.Run(":8000")
        }
Run and test:  

        $ go run hello.go
        $ curl http://localhost:8000/hello
        $ Hello, Wine!
        

## JSON Rendering

        s := wine.NewServer()
        s.Get("/time", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder {
        	return wine.JSON(map[string]interface{}{"time":time.Now().Unix()})
        })
        s.Run(":8000")

## Parameters
Request.Params() returns all parameters from URL query, post form, cookies, and custom header fields.

        s := wine.NewServer()
        s.Post("feedback", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder {
            text := req.Params().String("text")
            email := req.Params().String("email")
            return wine.Text("Feedback:" + text + " from " + email)
        })
        s.Run(":8000")
Support parameters in query string

        $ curl -X POST "http://localhost:8000/feedback?text=crash&email=wine@wine.com"
Support parameters in form

        $ curl -X POST -d "text=crash&email=wine@wine.com" http://localhost:8000/feedback
Support parameters in json

        $ curl -X POST -H "Content-Type:application/json" 
               -d '{"text":"crash", "email":"wine@wine.com"}' 
               http://localhost:8000/feedback
#### Parameters in URL Path
Path parameters are also supported in order to provide elegant RESTful apis.  
Single parameter in one segment:
<pre>
    s := wine.NewServer() 
    s.Get("/items/<b>{id}</b>", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder {
        id := req.Params().String("id")
        return wine.Text("item id: " + id)
    }) 
    s.Run(":8000")
</pre>
       
## Use Interceptor
Intercept and preprocess requests  

<pre>
    func Log(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder {
    	st := time.Now()  
    	//pass request to the next handler
    	<b>result := next(ctx, request)</b>
    	cost := float32((time.Since(st) / time.Microsecond)) / 1000.0
    	req := request.Request()
    	log.Printf("%.3fms %s %s", cost, req.Method, req.RequestURI)
    	return result
    } <br/>
    func main() {
    	s := wine.NewServer(nil) 
    	//Log every request
    	<b>r := s.Use(Log)</b> 
    	r.Get("/hello", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder {
    		return wine.Text("Hello, Wine!")
        })
        s.Run(":8000")
    }
</pre>
## Grouping Route
<pre>  
    func CheckSessionID(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder {
    	sid := req.Params().String("sid")
    	//check sid
    	if len(sid) == 0 {
    		return wine.JSON(map[string]interface{}{"error":"need sid"})
    	} else {
    		return next(ctx, request)
    	}
    }
    
    func GetUserProfile(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder  {
    	//...
    }
    
    func GetUserFriends(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder  {
    	//...
    }
    
    func GetServerTime(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder  {
    	//...
    }
    
    func main() {
    	s := wine.NewServer()
    
    	//Create "accounts" group and add interceptor CheckSessionID
    	<b>g := s.Group("accounts").Use(CheckSessionID)</b>
    	g.Get("{user_id}/profile", GetUserProfile)
    	g.Get("{user_id}/friends/{page}/{size}", GetUserFriends)
    
    	s.Get("time", GetServerTime)
    
    	s.Run(":8000")
    }  
</pre>
Run it: 

    Running at :8000 ...
    GET   /time/ main.GetServerTime
    GET   /accounts/{user_id}/friends/{page}/{size}    main.CheckSessionID, main.GetUserFriends
    GET   /accounts/{user_id}/profile/    main.CheckSessionID, main.GetUserProfile


## Auth
It's easy to turn on basic auth.

    s := wine.NewServer()
	s.Use(wine.NewBasicAuthHandler(map[string]string{
		"admin": "123",
		"tom":   "456",
	}, ""))
	s.StaticDir("/", "./html")
	s.Run(":8000")
	
## Recommendations
Wine designed for modular web applications/services is not a general purpose web server. It should be used behind a web server such as Nginx, Caddy which provide compression, security features.