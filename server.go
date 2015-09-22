package wine

type Server struct {
	*Router
}

/*
func (this *APIServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			Log().Error("ServeHTTP", e)
		}
		Log().Critical(fmt.Sprintf("Handled Request %q", req.RequestURI))
	}()

	Log().Critical(fmt.Sprintf("%s %s %q", req.Method, req.Header.Get(ContentTypeName), req.RequestURI))

	path := req.RequestURI
	i := strings.Index(path, "?")
	if i > 0 {
		path = req.RequestURI[:i]
	}

	handlers := this.router.Match(req.Method, path)
	if len(handlers) == 0 {
		Log().Error("No service ", path, "[", req.RequestURI, "]")
		http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	for _, h := range handlers {

	}

	if handler, parameters := self.findHandler(HTTPMethod(req.Method), path); handler != nil {
		logger.Info("Path Parameters: ", parameters)
		request := NewRequest(req, parameters)
		response := handler(request)
		if response != nil {
			//			logger.Info(path, response.Header, response.Status)
			for key, values := range response.Header {
				for _, v := range values {
					rw.Header().Set(key, v)
				}
			}
			rw.WriteHeader(response.Status)
			if len(response.Body) > 0 {
				rw.Write(response.Body)
			}
		} else {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			logger.Error(path, parameters)
		}
	} else {
	}
}*/