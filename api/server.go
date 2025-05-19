package main

import (
	"api/rest"
	"log"
	"net/http"
)

func ServeRest() {
	// rest without framework
	go func() {
		http.HandleFunc("/resthttp", rest.HandlerRestHttp())

		log.Println("REST without framework running at :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// rest with gin
	go func() {
		gin := rest.InitGin()
		gin.GET("/restgin", rest.HandlerGin())
		gin.Run(":8081")

	}()

	// rest with fiber
	go func() {
		fiber := rest.InitFiber()
		fiber.Get("/restfiber", rest.HandlerFiber())
		fiber.Listen(":8082")

	}()
}

func ServeGraphql() {

	// // graphql with gql
	// go func() {
	// 	h := handler.New(&handler.Config{
	// 		Schema:   &schema,
	// 		Pretty:   true,
	// 		GraphiQL: true,
	// 	})
	// 	http.Handle("/graphql", h)
	// 	log.Println("GraphQL server running at :8081")
	// 	log.Fatal(http.ListenAndServe(":8081", nil))
	// }()

}
