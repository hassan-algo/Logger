package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/examples/remote/msg"
	"github.com/anthdm/hollywood/remote"
)

type server struct {
}

func newServer() actor.Receiver {
	return &server{}
}

type LogMe struct {
	Method       string        `json:"Method"`
	URL          string        `json:"URL"`
	ResponseTime time.Duration `json:"ResponseTime"`
	Status       int           `json:"Status"`
	MicroService string        `json:"MicroService"`
}

func getLogFileName(address string) string {
	// Get the current date
	currentDate := time.Now().Format("2006-01-02") // Format: YYYY-MM-DD

	// Create a file name in the format: "address_YYYY-MM-DD.log"
	return fmt.Sprintf("%s_%s.log", address, currentDate)
}

func setupLogFile(address string) (*os.File, *log.Logger, error) {
	// Define the base directory for logs
	logDir := "server" // Directory where log files will be saved

	// Ensure the log directory exists
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, nil, fmt.Errorf("could not create log directory: %w", err)
	}

	// Get log file name based on address and current date
	logFileName := getLogFileName(address)

	// Correctly combine directory and file name using filepath.Join
	logFilePath := filepath.Join(logDir, logFileName)

	// Open the file in append mode, create if it doesn't exist
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open log file: %w", err)
	}

	// Create a new logger
	logger := log.New(file, "", log.LstdFlags)

	return file, logger, nil
}

func logMessage(address string, data interface{}, err error) {
	// Get the last part of the address for the log file name
	lastPart := filepath.Base(address)
	// Serialize the data object to JSON
	dataJson, _ := json.Marshal(data)

	// Serialize the error object to string if it's not nil
	var errStr string
	if err != nil {
		errStr = err.Error()
	} else {
		errStr = "nil"
	}

	// Setup the log file for the given address
	file, logger, err := setupLogFile(lastPart)
	if err != nil {
		log.Println("Error setting up log file:", err)
		return
	}
	defer file.Close()

	// Log the message
	logMessage := fmt.Sprintf("[ADDRESS]: %s\t [Data]: %s\t [Error]: %s", address, string(dataJson), errStr)
	logger.Println(logMessage)
}

func (f *server) Receive(ctx *actor.Context) {
	switch m := ctx.Message().(type) {
	case actor.Started:
		slog.Info("[SERVER STARTED]", "[ADDRESS]", ctx.PID().GetID())
	case *actor.PID:
		slog.Info("[CONNECTED]", "[ADDRESS]", ctx.PID().GetID(), "[PID]", m)
	case *msg.Message:
		data, err := decodeLogMe(m.Data)
		if err != nil {

			slog.Info("[ERROR]", "[ADDRESS]", ctx.PID().GetID(), "[ERROR]", err)
			logMessage(ctx.PID().GetID(), nil, err)
			break
		}
		slog.Info("[NEW]", "[ADDRESS]", ctx.PID().GetID(), "[MESSAGE]", data)
		logMessage(ctx.PID().GetID(), data, nil)
	default:
		slog.Warn("[UNKNOWN]", "[ADDRESS]", ctx.PID().GetID(), "[MESSAGE]", m, "[TYPE]", reflect.TypeOf(m).String())
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	r := remote.New("127.0.0.1:4002", remote.NewConfig())
	e, err := actor.NewEngine(actor.NewEngineConfig().WithRemote(r))
	if err != nil {
		panic(err)
	}

	e.Spawn(newServer, "server", actor.WithID("production"))
	e.Spawn(newServer, "server", actor.WithID("staging"))
	e.Spawn(newServer, "server", actor.WithID("dev"))
	select {}
}

func decodeLogMe(data string) (LogMe, error) {
	var logMe LogMe
	err := json.Unmarshal([]byte(data), &logMe)
	return logMe, err
}
