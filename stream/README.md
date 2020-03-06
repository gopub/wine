*Turn off reverse proxy buffering*  
It's suggested to turn off reverse proxy buffering in order to flush data to client immediately.   

    Nginx: proxy_buffering off
    Caddyserver: flush_interval -1