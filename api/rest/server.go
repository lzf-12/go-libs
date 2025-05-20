package rest

import (
	"log"
	"net/http"
)

func ServeRestHttp() error {

	http.HandleFunc("/resthttp", HandlerRestHttp())
	log.Println("REST without framework running at :8080")
	return http.ListenAndServe(":8080", nil)
}

func ServeRestGin() error {

	gin := InitGin()
	gin.GET("/restgin", HandlerGin())
	return gin.Run(":8081")
}

func ServeRestFiber() error {

	fiber := InitFiber()
	fiber.Get("/restfiber", HandlerFiber())
	return fiber.Listen(":8082")
}
