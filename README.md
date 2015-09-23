# wine
### A lightweight RESTful API Server implemented in golang
### Still in progress...

You can use wine like,   

		package main

		import (
			"github.com/justintan/wine"
			"github.com/justintan/gox"
			"time"
		)
		
		func main() {
			server := wine.Server()
		
			//add middleware
			server.Use(Logger)
		
			server.Get("/server_time", func (c *wine.Context)  {
				c.JSON(gox.M{"time":time.Now().Unix()})
			})
		
			server.Get("/topic/:id/title", func (c *wine.Context)  {
				topicId := c.Params.GetInt("id")
				gox.Log().Info(topicId)
				c.JSON(gox.M{"title":"this is topic title"})
			})
		
		
			//group router
			r := server.Group("/user")
			
			r.Get(":id/name", func (c *wine.Context)  {
				id := c.Params.GetInt("id")
				gox.Log().Info(id)
				c.JSON(gox.M{"name":"tom"})
			})
		
			r.Put(":id/name/:name", func (c *wine.Context)  {
				id := c.Params.GetInt("id")
				name := c.Params.GetStr("name")
				gox.Log().Info(id, name)
				c.JSON(gox.M{"name":name})
			})
		
			server.Run(":8080")
		}
		
		func Logger(c *wine.Context)  {
			gox.Log().Info("[BEGIN]", c.Request.URL.RequestURI())
			c.Next()
			gox.Log().Info("[END]", c.Request.URL.RequestURI())
		}
