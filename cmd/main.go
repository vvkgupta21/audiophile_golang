package main

import (
	"audio_phile/database"
	"audio_phile/server"
	"github.com/sirupsen/logrus"
)

func main() {
	srv := server.SetupRoutes()
	if err := database.ConnectAndMigrate(
		"localhost",
		"5434",
		"postgres",
		"local",
		"local",
		database.SSLModeDisable); err != nil {
		logrus.Panicf("Failed to initialize and migrate database with error: %+v", err)
	}
	logrus.Info("migration successfully!!")

	if err := srv.Run(":8000"); err != nil {
		logrus.Fatalf("Failed to run server with error %+v", err)
	}
	logrus.Print("Server started at : 8000")
}
