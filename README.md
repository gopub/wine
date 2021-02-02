# WINE

Wine is a lightweight web framework for quickly writing web applications/services in Go. 

## Install  

        $ go get -u -v github.com/gopub/wine

## Quick start  
Create ./hello.go  
        
        package main
        
        import "github.com/gopub/wine"
        
        func main() {
        	s := wine.NewServer(nil)
        	s.Get("/hello", func(ctx context.Context, req *wine.Request) wine.Responder {
        		return wine.Text("Hello, Wine!")
        	})
        	s.Run(":8000")
        }
Run and test:  

        $ go run hello.go
        $ curl http://localhost:8000/hello
        $ Hello, Wine!
        

## JSON Rendering

        s := wine.NewServer(nil)
        s.Get("/time", func(ctx context.Context, req *wine.Request) wine.Responder {
        	return wine.JSON(http.StatusOK, map[string]interface{}{"time":time.Now().Unix()})
        })
        s.Run(":8000")

## Parameters
Request.Params() returns all parameters from URL query, post form, cookies, and custom header fields.

        s := wine.NewServer(nil)
        s.Post("feedback", func(ctx context.Context, req *wine.Request) wine.Responder {
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
Path parameters are also supported in order to provide elegant RESTFul API.  
Single parameter in one segment:
<pre>
    s := wine.NewServer(nil) 
    s.Get("/items/<b>{id}</b>", func(ctx context.Context, req *wine.Request) wine.Responder {
        id := req.Params().String("id")
        return wine.Text("item id: " + id)
    }) 
    s.Run(":8000")
</pre>

## Model Binding
If an endpoint is bound with a model, request's parameters will be unmarshalled into an instance of the same model type. <br>
If the model implements interface wine.Validator, then it will be validated before passed to handler.
<pre>
type Item struct {
    ID      string      `json:"id"`
    Price   float32     `json:"price"`
}
func (i *Item) Validate() error {
    if i.ID == "" {
        return errors.BadRequest("missing id")
    }
    if i.Price <= 0 {
        return errors.BadRequest("missing or invalid price")
    }
    return nil 
}

func main() {
    s := wine.NewServer(nil) 
    s.Post("/items", CreateItem).SetModel(&Item{}) 
    s.Run(":8000")
}
func CreateItem(ctx context.Context, req *wine.Request) wine.Responder {
    // It's safe to get *Item, and item has been validated
    item := req.Model.(*Item)
    // TODO: save item into database
    return item
}
</pre>
       
## Use Interceptor
Intercept and preprocess requests  

<pre>
    func Log(ctx context.Context, req *wine.Request) wine.Responder {
    	st := time.Now()  
    	//pass request to the next handler
    	<b>result := wine.Next(ctx, request)</b>
    	cost := float32((time.Since(st) / time.Microsecond)) / 1000.0
    	req := request.Request()
    	log.Printf("%.3fms %s %s", cost, req.Method, req.RequestURI)
    	return result
    } <br/>
    func main() {
    	s := wine.NewServer(nil) 
    	//Log every request
    	<b>r := s.Use(Log)</b> 
    	r.Get("/hello", func(ctx context.Context, req *wine.Request) wine.Responder {
    		return wine.Text("Hello, Wine!")
        })
        s.Run(":8000")
    }
</pre>
## Grouping Route
<pre>  
    func CheckSessionID(ctx context.Context, req *wine.Request) wine.Responder {
    	sid := req.Params().String("sid")
    	//check sid
    	if len(sid) == 0 {
    		return errors.BadRequest("missing sid")
    	} 
    	return wine.Next(ctx, request)
    }
    
    func GetUserProfile(ctx context.Context, req *wine.Request) wine.Responder  {
    	//...
    }
    
    func GetUserFriends(ctx context.Context, req *wine.Request) wine.Responder  {
    	//...
    }
    
    func GetServerTime(ctx context.Context, req *wine.Request) wine.Responder  {
    	//...
    }
    
    func main() {
    	s := wine.NewServer(nil)
    
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

    s := wine.NewServer(nil)
	s.Use(wine.NewBasicAuthHandler(map[string]string{
		"admin": "123",
		"tom":   "456",
	}, ""))
	s.StaticDir("/", "./html")
	s.Run(":8000")
	
## Recommendations
Wine designed for modular web applications/services is not a general purpose web server. It should be used behind a web server such as Nginx, Caddy which provide compression, security features.