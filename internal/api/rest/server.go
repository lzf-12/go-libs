package rest

import (
	"context"
	"log"
	"net/http"
)

func ServeRestHttp(ctx context.Context) error {

	http.HandleFunc("/resthttp", HandlerRestHttp())
	log.Println("REST without framework running at :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		return err
	}

	<-ctx.Done()
	// handle dependency shutdown here

	return nil
}

func ServeRestGin(ctx context.Context) error {

	gin := InitGin()
	gin.GET("/restgin", HandlerGin())
	if err := gin.Run(":8081"); err != nil {
		return err
	}

	<-ctx.Done()
	// handle dependency shutdown here

	return nil
}

func ServeRestFiber(ctx context.Context) error {

	fiber := InitFiber()
	fiber.Get("/restfiber", HandlerFiber())
	if err := fiber.Listen(":8082"); err != nil {
		return err
	}

	<-ctx.Done()
	// handle dependency shutdown here

	return nil
}
