This package is to build a virtual file system supporting http access.

* FileSystem implements http.FileSystem  
* File implements http.File  
* FileInfo implements os.FileInfo  

Known issue:
1. Cannot stream large video/quicktime files for AVPlayer to play. Playing small size video file looks good. Downloading large video files works well.