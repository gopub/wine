# wine
### A lightweight RESTful API Server implemented in golang
### Still in progress...

You can use wine like,   

		
		func main() {
        
        	s := wine.Server()
        
        	//You can implement middle wares and add them to the routing
        	s.Use(wine.Logger)
        
        
        	//support file and directory
        	s.StaticFile("/", "/var/www/index.html")
        	s.StaticDir("/html/*", "/var/www/html")
        
        	s.GET("server-time", func(c wine.Context) {
        		resp := map[string]interface{}{"time": time.Now().Unix()}
        		c.SendJSON(resp)
        	})
        
        	s.GET("users/:id/name", func(c wine.Context) {
        		id := c.RequestParams().GetStr("id")
        		resp := map[string]interface{}{"name": "This is " + id + "'s name"}
        		c.SendJSON(resp)
        	})
        
        	//Equals to s.GET("login", login) and s.POST("login", login)
        	s.GP("login", login)
        
        	g := s.Group("users")
        	g.Use(auth)
        	g.GET(":id/profile", getProfile)
        	g.PUT(":id/name", updateName)
        
        	s.Run(":8080")
        }
        
        func auth(c wine.Context) {
        	sid := c.Get("session_id")
        	fmt.Println(sid)
        	//auth sid
        	//...
        	authorized := false
        
        	if authorized {
        		//call the next handler
        		c.Next()
        	} else {
        		//abort the handling process, send an error response
        		resp := map[string]interface{}{"msg": "authorization failed"}
        		c.SendJSON(resp)
        	}
        }
        
        func login(c wine.Context) {
        	account := c.RequestParams().GetStr("account")
        	password := c.RequestParams().GetStr("password")
        	fmt.Println(account, password)
        	resp := map[string]interface{}{"status": "success"}
        	c.SendJSON(resp)
        }
        
        func getProfile(c wine.Context) {
        	id := c.RequestParams().GetStr("id")
        	resp := map[string]interface{}{"profile": "This is " + id + "'s profile"}
        	c.SendJSON(resp)
        }
        
        func updateName(c wine.Context) {
        	name := c.RequestParams().GetStr("name")
        	resp := map[string]interface{}{"name": name}
        	c.SendJSON(resp)
        }



2015/10/04 20:36:25 [INFO ]  Running at :8080 ...  
2015/10/04 20:36:25 [INFO ] GET   /     github.com/justintan/wine.(*Router).StaticFile.func1  
2015/10/04 20:36:25 [INFO ] GET   /users/:id/profile    github.com/justintan/wine.Logger, main.auth, main.getProfile  
2015/10/04 20:36:25 [INFO ] GET   /login        github.com/justintan/wine.Logger, main.login  
2015/10/04 20:36:25 [INFO ] GET   /users/:id/name       github.com/justintan/wine.Logger, main.main.func2  
2015/10/04 20:36:25 [INFO ] GET   /server-time  github.com/justintan/wine.Logger, main.main.func1  
2015/10/04 20:36:25 [INFO ] GET   /html/*       github.com/justintan/wine.Logger, github.com/justintan/wine.(*Router).StaticFS.func1  
2015/10/04 20:36:25 [INFO ] POST  /login        github.com/justintan/wine.Logger, main.login  
2015/10/04 20:36:25 [INFO ] PUT   /users/:id/name       github.com/justintan/wine.Logger, main.auth, main.updateName  
  