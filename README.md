# wine
### A lightweight RESTful API Server implemented in golang
### Still in progress...

You can use wine like,   

1. Webpage server  
		
		s := wine.Default()
        s.StaticDir("/", "./html")
        s.Run(":8000")
        
2. RESTFul API Server  

        s := wine.Default()
    
    	s.Get("whattime", func(c wine.Context) {
    		c.JSON(types.M{"time": time.Now()})
    	})
    
    	s.Get("users/:user_id/name", func(c wine.Context) {
    		c.HTML(c.Params().GetStr("user_id") + "'s name is Wine")
    	})
    
    	s.Post("users/:user_id/name/:name", func(c wine.Context) {
    		c.HTML(c.Params().GetStr("user_id") + "'s new name is " + c.Params().GetStr("name"))
    	})
    
    	s.Any("login", func(c wine.Context) {
    		username := c.Params().GetStr("username")
    		password := c.Params().GetStr("password")
    		fmt.Println(username, password)
    		c.JSON(types.M{"status": 0, "token": types.NewUUID(), "msg": "success"})
    	})
    
    	s.Run(":8000")
        	
 
  