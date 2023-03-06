package main

import (
	"audio_phile/database"
	"audio_phile/model"
	"audio_phile/server"
	"audio_phile/utils"
	"github.com/sirupsen/logrus"
)

func main() {
	srv := server.SetupRoutes()
	err := database.ConnectAndMigrate(
		"localhost",
		"5434",
		"postgres",
		"local",
		"local",
		database.SSLModeDisable)
	if err != nil {
		logrus.Panicf("Failed to initialize and migrate database with error: %+v", err)
	}
	logrus.Info("migration successfully!!")

	model.FirebaseClient, err = utils.GetFirebaseClient()
	if err != nil {
		logrus.Errorf("Firebase client createtion failed; %s", err.Error())
		return
	}

	if err := srv.Run(":8000"); err != nil {
		logrus.Fatalf("Failed to run server with error %+v", err)
	}
	logrus.Print("Server started at : 8000")
}
