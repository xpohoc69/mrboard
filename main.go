package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/xpohoc69/mrboard/models"
	"github.com/xpohoc69/mrboard/requesters"
	"github.com/xpohoc69/mrboard/services"
)

var flags = &models.Flags{}
var config = &models.Config{}
var requester = requesters.NewRequester(config)
var service = services.NewMrService(config, requester, flags)

func init() {
	log.SetFlags(log.Lshortfile)

	flag.StringVar(&flags.Env, "env", "", "Path to .env file. Default is current directory")
	flag.BoolVar(&flags.OnlyMine, "mine", false, "Show merge requests where I am the author. Default is false")
	flag.BoolVar(&flags.NeedMyApprove, "nma", false, "Show merge requests where my approval is needed. Default is false")
	flag.Parse()

	envPath := ".env"
	if flags.Env != "" {
		envPath = flags.Env
	}

	if err := godotenv.Load(envPath); err != nil {
		log.Fatal("Error loading .env file")
	}
	splitUsers := strings.Split(os.Getenv("APP_USERS"), ",")
	users := make(map[string]string, 8)
	for _, user := range splitUsers {
		users[user] = user
	}
	config.Me = os.Getenv("APP_ME")
	config.Users = users
	config.ApiToken = os.Getenv("APP_API_TOKEN")
	config.ApiUrl = os.Getenv("APP_GITLAB_API_URL")
	config.ProjectId = os.Getenv("APP_PROJECT_ID")
	config.TaskUrl = os.Getenv("APP_TASK_URL")
	config.TaskIdRegex = os.Getenv("APP_TASK_ID_REGEX")
}

func main() {
	result := service.PrepareResult()
	service.PrintTable(result)
}
