package utils

import (
	"audio_phile/model"
	cloud "cloud.google.com/go/storage"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go/v4"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/option"
	"io"
	"net/http"
	"os"
	"strings"
)

type clientError struct {
	ID            string `json:"id"`
	MessageToUser string `json:"messageToUser"`
	DeveloperInfo string `json:"developerInfo"`
	Err           string `json:"error"`
	StatusCode    int    `json:"statusCode"`
	IsClientError bool   `json:"isClientError"`
}

func ParseBody(body io.Reader, out interface{}) error {
	err := json.NewDecoder(body).Decode(out)
	if err != nil {
		return err
	}
	return nil
}

func RespondError(w http.ResponseWriter, statusCode int, err error, messageToUser string, additionalInfoForDevs ...string) {
	logrus.Errorf("status: %d, message: %s, err: %+v ", statusCode, messageToUser, err)
	clientError := newClientError(statusCode, err, messageToUser, additionalInfoForDevs...)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(clientError); err != nil {
		logrus.Errorf("Failed to send error to caller with error: %+v", err)
	}
}

func newClientError(statusCode int, err error, messageToUser string, additionalInfoForDevs ...string) *clientError {
	additionalInfoJoined := strings.Join(additionalInfoForDevs, "\n")
	if additionalInfoJoined == "" {
		additionalInfoJoined = messageToUser
	}

	errorId, _ := shortid.Generate()
	var errString string
	if err != nil {
		errString = err.Error()
	}

	return &clientError{
		ID:            errorId,
		MessageToUser: messageToUser,
		DeveloperInfo: additionalInfoJoined,
		Err:           errString,
		StatusCode:    statusCode,
		IsClientError: true,
	}
}

// EncodeJsonBody write the JSON body to response writer
func EncodeJsonBody(w http.ResponseWriter, body interface{}) error {
	return json.NewEncoder(w).Encode(body)
}

// RespondJSON sends the interface as a json
func RespondJSON(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if body != nil {
		if err := EncodeJsonBody(w, body); err != nil {
			logrus.Errorf("failed to respond JSON with error: %+v", err)
		}
	}
}

// EncodeJSONBody writes the JSON body to response writer
func EncodeJSONBody(resp http.ResponseWriter, data interface{}) error {
	return json.NewEncoder(resp).Encode(data)
}

// HashPassword returns the bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func CheckPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func GetFirebaseClient() (*model.App, error) {
	client := &model.App{}
	client.Ctx = context.Background()
	credentialsFile := option.WithCredentialsJSON([]byte(GetEnvValue("Firebase_Storage_Credential")))
	app, err := firebase.NewApp(client.Ctx, nil, credentialsFile)
	if err != nil {
		return client, err
	}

	client.Client, err = app.Firestore(client.Ctx)
	if err != nil {
		return client, err
	}

	client.Storage, err = cloud.NewClient(client.Ctx, credentialsFile)
	if err != nil {
		return client, err
	}

	return client, nil
}

func LoadEnv(filename string) error {
	err := godotenv.Load(filename)
	if err != nil {
		return err
	}
	return nil
}

func GetEnvValue(key string) string {
	value := os.Getenv(key)
	return value
}
