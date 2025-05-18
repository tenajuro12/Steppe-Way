package main

import (
	"context"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"plan_service/internal/handlers"
	"plan_service/internal/models"
	database "plan_service/utils/db"
	"time"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(
		&models.Plan{},
		&models.PlanItem{},
		&models.PlanTemplate{},
		&models.TemplateItem{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database schemas: %v", err)
	}

	r := mux.NewRouter()

	planHandler := handlers.PlanHandler{}

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/plans", planHandler.GetUserPlans).Methods("GET")
	api.HandleFunc("/plans", planHandler.CreatePlan).Methods("POST")
	api.HandleFunc("/plans/{id:[0-9]+}", planHandler.GetPlan).Methods("GET")
	api.HandleFunc("/plans/{id:[0-9]+}", planHandler.UpdatePlan).Methods("PUT")
	api.HandleFunc("/plans/{id:[0-9]+}", planHandler.DeletePlan).Methods("DELETE")
	api.HandleFunc("/plans/{id:[0-9]+}/items", planHandler.AddItemToPlan).Methods("POST")
	api.HandleFunc("/plans/{id:[0-9]+}/optimize", planHandler.OptimizeRoute).Methods("POST")
	api.HandleFunc("/plans/items/{itemId:[0-9]+}", planHandler.UpdatePlanItem).Methods("PUT")
	api.HandleFunc("/plans/items/{itemId:[0-9]+}", planHandler.DeletePlanItem).Methods("DELETE")

	api.HandleFunc("/plans/{id:[0-9]+}/directions", planHandler.GetPlanDirections).Methods("GET")
	api.HandleFunc("/plans/items/{fromItemId:[0-9]+}/directions/{toItemId:[0-9]+}", planHandler.GetDirectionsBetweenItems).Methods("GET")

	api.HandleFunc("/templates", planHandler.GetTemplates).Methods("GET")
	api.HandleFunc("/templates", planHandler.CreateTemplate).Methods("POST")
	api.HandleFunc("/templates/{id:[0-9]+}", planHandler.GetTemplate).Methods("GET")
	api.HandleFunc("/templates/{id:[0-9]+}", planHandler.UpdateTemplate).Methods("PUT")
	api.HandleFunc("/templates/{id:[0-9]+}", planHandler.DeleteTemplate).Methods("DELETE")
	api.HandleFunc("/templates/{id:[0-9]+}/items", planHandler.GetTemplateItems).Methods("GET")
	api.HandleFunc("/templates/{id:[0-9]+}/items", planHandler.AddItemToTemplate).Methods("POST")
	api.HandleFunc("/templates/items/{itemId:[0-9]+}", planHandler.UpdateTemplateItem).Methods("PUT")
	api.HandleFunc("/templates/items/{itemId:[0-9]+}", planHandler.DeleteTemplateItem).Methods("DELETE")
	api.HandleFunc("/templates/create-plan", planHandler.CreatePlanFromTemplate).Methods("POST")

	srv := &http.Server{
		Addr:         ":8087",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		log.Println("Starting plan service on port 8087...")
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("Server gracefully stopped")
	os.Exit(0)
}
