package rest

import (
	"log"
	"net/http"
)

func ServeRest() {
	// without framework
	go func() {
		http.HandleFunc("/resthttp", HandlerRestHttp())

		log.Println("REST without framework running at :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// gin
	go func() {
		gin := InitGin()
		gin.GET("/restgin", HandlerGin())
		gin.Run(":8081")

	}()

	// fiber
	go func() {
		fiber := InitFiber()
		fiber.Get("/restfiber", HandlerFiber())
		fiber.Listen(":8082")

	}()
}
