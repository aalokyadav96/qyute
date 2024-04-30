package main

import (
    "net/http"
    "log"

    "github.com/julienschmidt/httprouter"
)


const maxUploadSize = 10 * 1024 * 1024 // 10 mb
const maxImageUploadSize = 1 * 1024 * 1024 // 1 mb

func main() {
	router := httprouter.New()
	router.POST("/up", (UploadVideoFileHandler))
	router.GET("/image/:imageName", sendImageAsBytes)
	router.GET("/csrf", CSRF)
	router.POST("/evt", Res)
	router.GET("/evt", Res)
	router.POST("/res", Res)
	router.GET("/res", Res)
	router.POST("/translate", Translate)
	router.GET("/fav/favicon.ico", Ignore)
	static := httprouter.New()
	//~ static.ServeFiles("/video/*filepath", http.Dir("./videos"))
	static.ServeFiles("/video/*filepath", http.Dir("./streams"))
	//~ static.ServeFiles("/poster/*filepath", http.Dir(postersDir))
	//static.ServeFiles("/image/*filepath", http.Dir("./images"))
	router.ServeFiles("/static/*filepath", http.Dir("static"))

//	router.NotFound = http.FileServer(http.Dir(""))
	router.ServeFiles("/usrimg/*filepath", http.Dir("usrimg"))
	router.NotFound = static


	log.Println("Starting Server")
    log.Fatal(http.ListenAndServe(":4371", router))
}
