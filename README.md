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
Context.Params() provides an uniform interface to retrieve request parameters.  

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

        $ curl -X POST -H "Content-Type:application/json" 
               -d '{"text":"crash", "email":"wine@wine.com"}' 
               http://localhost:8000/feedback
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
## Grouping Route
<pre>  
    func CheckSessionID(c wine.Context) {
    	sid := c.Params().GetStr("sid")
    	//check sid
    	if len(sid) == 0 {
    		c.JSON(map[string]interface{}{"error":"need sid"})
    	} else {
    		c.Next()
    	}
    }
    
    func GetUserProfile(c wine.Context)  {
    	//...
    }
    
    func GetUserFriends(c wine.Context)  {
    	//...
    }
    
    func GetServerTime(c wine.Context)  {
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

    [WINE] Running at :8000 ...
    [WINE] GET   /time/ main.GetServerTime
    [WINE] GET   /accounts/:user_id/friends/:page,:size/    main.CheckSessionID, main.GetUserFriends
    [WINE] GET   /accounts/:user_id/profile/    main.CheckSessionID, main.GetUserProfile

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
    	s := wine.Default()
    	s.Post("register", func(c wine.Context) {
    		u := &User{}
    		c.Params().AssignTo(u, "param")
    		c.JSON(u)
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
## Custom Context
Custom context to add more features. 
e.g. Create MyContext to support SendResponse method 

    type MyContext struct {
    	*wine.DefaultContext
    	handlers *wine.HandlerChain
    }
    
    func (c *MyContext) Rebuild(rw http.ResponseWriter, req *http.Request, templates []*template.Template, handlers []wine.Handler) {
    	if c.DefaultContext == nil {
    		c.DefaultContext = &wine.DefaultContext{}
    	}
    	c.DefaultContext.Rebuild(rw, req, templates, handlers)
    	c.handlers = wine.NewHandlerChain(handlers)
    }
    
    func (c *MyContext) Next() {
    	if h := c.handlers.Next(); h != nil {
    		h.HandleRequest(c)
    	}
    }
    
    func (c *MyContext) SendResponse(code int, msg string, data interface{}) {
    	c.JSON(map[string]interface{}{"code": code, "data": data, "msg": msg})
    }
    
    
    func main() {
    	s := wine.Default()
    	s.RegisterContext(&MyContext{})
    	s.Get("time", func(c wine.Context) {
    		ctx := c.(*MyContext)
    		ctx.SendResponse(0, "", time.Now().Unix())
    	})
    	s.Run(":8000")
    }
Test:  

    $ curl http://localhost:8000/time
Response:  

    {
        "code": 0,
        "data": 1459404100,
        "msg": ""
    }    
## Recommendations
    Wine which is designed for modular web applications/services is not a general purpose web server. It should be used behind a web server such as Nginx, Caddy which provide compression, security features.