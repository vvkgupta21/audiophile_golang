package main

import (
	"audio_phile/database"
	"audio_phile/model"
	"audio_phile/server"
	"audio_phile/utils"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"time"
)

func main() {

	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt)

	err := utils.LoadEnv(".env")
	if err != nil {
		logrus.Errorf("Environment variables loading failed.; %s", err.Error())
		return
	}

	model.FirebaseClient, err = utils.GetFirebaseClient()
	if err != nil {
		logrus.Errorf("Firebase client createtion failed; %s", err.Error())
		return
	}

	//port := utils.GetEnvValue("DB_Port")
	srv := server.SetupRoutes()
	err = database.ConnectAndMigrate()
	if err != nil {
		return
	}

	go func() {
		if err := srv.Run(":8000"); err != nil {
			logrus.Errorf("Server not shut down gracefully; %s", err.Error())
			return
		}
	}()

	<-done
	database.CloseDb()
	if err := srv.Stop(5 * time.Second); err != nil {
		logrus.Errorf("Server not shut down gracefully; %s", err.Error())
		return
	}
}
