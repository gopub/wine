# wine
### A lightweight RESTful API Server implemented in golang
### Still in progress...

You can use wine like,   

		
		func main() {
        	s := wine.Server()
        
        	//You can implement middle wares and add them to the routing
        	s.Use(logger)
        
        	s.GET("server-time", func(c *wine.Context) {
        		resp := map[string]interface{}{"time": time.Now().Unix()}
        		c.JSON(resp)
        	})
        
        	s.GET("users/:id/name", func(c *wine.Context) {
        		id := c.RequestParams.GetStr("id")
        		resp := map[string]interface{}{"name": "This is " + id + "'s name"}
        		c.JSON(resp)
        	})
        
        	//ANY means the union of GET, POST, PUT, DELETE
        	s.ANY("login", login)
        
        	g := s.Group("users")
        	g.Use(auth)
        	g.GET(":id/profile", getProfile)
        	g.PUT(":id/name", updateName)
        
        	s.Run(":8080")
        }
        
        func auth(c *wine.Context) {
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
        		c.JSON(resp)
        	}
        }
        
        func logger(c *wine.Context)  {
        	startAt := time.Now()
        	c.Next()
        	duration := time.Since(startAt)
        	timeStr := fmt.Sprintf("%dus", duration/1000)
        	gox.LInfo(c.Request.Method, c.Request.RequestURI, timeStr)
        }
        
        func login(c *wine.Context) {
        	account := c.RequestParams.GetStr("account")
        	password := c.RequestParams.GetStr("password")
        	fmt.Println(account, password)
        	resp := map[string]interface{}{"status": "success"}
        	c.JSON(resp)
        }
        
        func getProfile(c *wine.Context) {
        	id := c.RequestParams.GetStr("id")
        	resp := map[string]interface{}{"profile": "This is " + id + "'s profile"}
        	c.JSON(resp)
        }
        
        func updateName(c *wine.Context) {
        	name := c.RequestParams.GetStr("name")
        	resp := map[string]interface{}{"name": name}
        	c.JSON(resp)
        }


2015/09/27 21:52:48 [INFO ]  Running at :8080 ...  
2015/09/27 21:52:48 [INFO ] GET   /users/:id/profile    main.auth, main.getProfile  
2015/09/27 21:52:48 [INFO ] GET   /login        main.logger, main.login  
2015/09/27 21:52:48 [INFO ] GET   /users/:id/name       main.logger, main.main.func2  
2015/09/27 21:52:48 [INFO ] GET   /server-time  main.logger, main.main.func1  
2015/09/27 21:52:48 [INFO ] POST  /login        main.logger, main.login  
2015/09/27 21:52:48 [INFO ] PUT   /users/:id/name       main.auth, main.updateName  
2015/09/27 21:52:48 [INFO ] PUT   /login        main.logger, main.login  
2015/09/27 21:52:48 [INFO ] DELETE /login       main.logger, main.login  
2015/09/27 21:52:56 [CRITI]  GET  "/server-time"  
2015/09/27 21:52:56 [INFO ]  GET /server-time 115us  
2015/09/27 21:55:52 [CRITI]  POST  "/login?account=tom&password=123"  
tom 123  
2015/09/27 21:55:52 [INFO ]  POST /login?account=tom&password=123 37us  