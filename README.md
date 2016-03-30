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
Single parameter in a path segment

        s := wine.Default()
        s.Get("/items/:id", func(c wine.Context) {
        	id := c.Params().GetStr("id")
        	c.Text("item id: " + id)
        })
        s.Run(":8000")
        
Multiple parameters in a path segment   
     
        s.Get("/items/:page,:size", func(c wine.Context) {
        	page := c.Params().GetInt("page")
        	size := c.Params().GetInt("size")
        	c.Text("page:" + strconv.Itoa(page) + " size:" + strconv.Itoa(size))
        })

        	
 
  