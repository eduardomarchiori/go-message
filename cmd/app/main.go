package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/eduardomarchiori/go-api/internal/infra/akafka"
	"github.com/eduardomarchiori/go-api/internal/infra/repository"
	"github.com/eduardomarchiori/go-api/internal/infra/web"
	"github.com/eduardomarchiori/go-api/internal/usecase"
	"github.com/go-chi/chi/v5"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(host.docker.internal:3306)/products")

	if err != nil {
		panic(err)
	}
	defer db.Close()

	repository := repository.NewProductRepositoryMYsql(db)
	createProductUsecase := usecase.NewCreateProductUseCase(repository)
	listProductUseCase := usecase.NewListProductsUseCase(repository)

	productHandlers := web.NewProductHandlers(
		createProductUsecase,
		listProductUseCase,
	)

	r := chi.NewRouter()
	r.Post("/products", productHandlers.CreateProductHandler)
	r.Get("/products", productHandlers.ListProductHandler)

	go http.ListenAndServe(":8000", r)

	msgChan := make(chan *kafka.Message)
	go akafka.Consume([]string{"products"}, "host.docker.internal:9092", msgChan)

	for msg := range msgChan {
		dto := usecase.CreateProductInputDTO{}
		err := json.Unmarshal(msg.Value, &dto)
		if err != nil {
			println(err)
		}

		_, err = createProductUsecase.Execute(dto)
	}

}
