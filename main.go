package main

import (
	"log"
	"os"
	"strings"
	"net/http"
	"fmt"

	"github.com/google/uuid"
)

var (
	DATA_PATH = getenv("DATA_PATH", "/var/staticinator")
	PORT = getenv("PORT", "7878")
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "create":
			host := os.Args[2]
			directory := getDirectory(host)

			createDirs(directory + "/html")

			token := uuid.NewString()
			os.WriteFile(directory + "/token", []byte(token), 0600)

			fmt.Println("Created '" + directory + "'. Your token is:")
			fmt.Println(token)		
		}
	} else {
		serve()
	}
}

func serve() {
	http.Handle("/update", updateHandler())
	log.Fatal(http.ListenAndServe(":" + PORT, nil))
}

func getDirectory(host string) string {
	return DATA_PATH + "/" + strings.ReplaceAll(host, "/", "-");
}

func createDirs(dirs string) {
	err := os.MkdirAll(dirs, 0755)
	if err != nil {
		log.Panic(err)
	}
}

func getenv(key string, defaultVal string) string {
    if value, exists := os.LookupEnv(key); exists {
		return value
    }

    return defaultVal
}