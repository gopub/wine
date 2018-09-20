# WINE

Wine is a lightweight web framework for quickly writing web applications/services in Go. 

## Install  

        $ go get -u -v github.com/gopub/wine

## Quick start  
Create ./hello.go  
        
        package main
        
        import "github.com/gopub/wine/v3"
        
        func main() {
        	s := wine.DefaultServer()
        	s.Get("/hello", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
        		return wine.Text("Hello, Wine!")
        	})
        	s.Run(":8000")
        }
Run and test:  

        $ go run hello.go
        $ curl http://localhost:8000/hello
        $ Hello, Wine!
        

## JSON Rendering

        s := wine.DefaultServer()
        s.Get("/time", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
        	return wine.JSON(map[string]interface{}{"time":time.Now().Unix()})
        })
        s.Run(":8000")

## Parameters
Context.Params() provides an uniform interface to retrieve request parameters.  

        s := wine.DefaultServer()
        s.Post("feedback", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
            text := request.Parameters().String("text")
            email := request.Parameters().String("email")
            return wine.Text("Feedback:" + text + " from " + email)
        })
        s.Run(":8000")
Test parameters in query string

        $ curl -X POST "http://localhost:8000/feedback?text=crash&email=wine@wine.com"
Test parameters in form

        $ curl -X POST -d "text=crash&email=wine@wine.com" http://localhost:8000/feedback
Test parameters in json

        $ curl -X POST -H "Content-Type:application/json" 
               -d '{"text":"crash", "email":"wine@wine.com"}' 
               http://localhost:8000/feedback
#### Parameters in URL Path
Path parameters are also supported in order to provide elegant RESTful apis.  
Single parameter in one segment:
<pre>
    s := wine.DefaultServer() 
    s.Get("/items/<b>:id</b>", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
        id := request.Parameters().String("id")
        return wine.Text("item id: " + id)
    }) 
    s.Run(":8000")
</pre>
        
Multiple parameters in one segment:   
<pre>
    s := wine.DefaultServer() 
    s.Get("/items/<b>:page,:size</b>", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
        page := request.Parameters().Int("page")
        size := request.Parameters().Int("size")
        return wine.Text("page:" + strconv.Itoa(page) + " size:" + strconv.Itoa(size))
    }) 
    s.Run(":8000")
</pre>

## Use Middlewares
Use middlewares to intercept and preprocess requests  

Custom middleware
<pre>
    func Logger(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
    	st := time.Now()  
    	//pass request to the next handler
    	<b>result := invoker(ctx, request)</b>
    	cost := float32((time.Since(st) / time.Microsecond)) / 1000.0
    	req := return wine.Request()
    	log.Printf("%.3fms %s %s", cost, req.Method, req.RequestURI)
    	return result
    } <br/>
    func main() {
    	s := wine.NewServer() 
    	//Use middleware Logger
    	<b>s.Use(Logger)</b> 
    	s.Get("/hello", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
    		return wine.Text("Hello, Wine!")
        })
        s.Run(":8000")
    }
</pre>
## Grouping Route
<pre>  
    func CheckSessionID(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
    	sid := request.Parameters().String("sid")
    	//check sid
    	if len(sid) == 0 {
    		return wine.JSON(map[string]interface{}{"error":"need sid"})
    	} else {
    		return invoker(ctx, request)
    	}
    }
    
    func GetUserProfile(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible  {
    	//...
    }
    
    func GetUserFriends(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible  {
    	//...
    }
    
    func GetServerTime(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible  {
    	//...
    }
    
    func main() {
    	s := wine.NewServer()
    
    	//Create "accounts" group
    	<b>g := s.Group("accounts")</b>
    	//Use CheckSessionID to process all requests in this route group
    	<b>g.Use(CheckSessionID)</b>
    	g.Get(":user_id/profile", GetUserProfile)
    	g.Get(":user_id/friends/:page,:size", GetUserFriends)
    
    	s.Get("time", GetServerTime)
    
    	s.Run(":8000")
    }  
</pre>
Run it: 

    Running at :8000 ...
    GET   /time/ main.GetServerTime
    GET   /accounts/:user_id/friends/:page,:size/    main.CheckSessionID, main.GetUserFriends
    GET   /accounts/:user_id/profile/    main.CheckSessionID, main.GetUserProfile

## Model Binding

    type Coordinate struct {
    	Lat float64 `json:"lat" param:"lat"`
    	Lng float64 `json:"lng" param:"lng"`
    }
    
    type User struct {
    	ID         int         `json:"id" param:"-"`
    	Name       string      `json:"name" param:"name"`
    	Password   string      `json:"-" param:"password"`
    	Coordinate *Coordinate `json:"coordinate" param:"coordinate"`
    }
    
    func main() {
    	s := wine.DefaultServer()
    	s.Post("register", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
    		u := &User{}
    		request.Parameters().AssignTo(u, "param")
    		return wine.JSON(u)
    	})
    	s.Run(":8000")
    }
Test:
    
    $ curl -X POST -H "Content-Type:application/json" 
           -d '{"name":"tom", "password":"123", "coordinate":{"lat":21, "lng":90.0}}' 
           http://localhost:8000/register
Response:

    {
         "id": 1,
         "name": "tom",
         "coordinate": {
             "lat": 21,
             "lng": 90
         }
    }
## Basic Auth
It's easy to turn on basic auth.

    s := wine.DefaultServer()
	s.Use(wine.BasicAuth(map[string]string{
		"admin": "123",
		"tom":   "456",
	}, ""))
	s.StaticDir("/", "./html")
	s.Run(":8000")

## Recommendations
Wine designed for modular web applications/services is not a general purpose web server. It should be used behind a web server such as Nginx, Caddy which provide compression, security features.