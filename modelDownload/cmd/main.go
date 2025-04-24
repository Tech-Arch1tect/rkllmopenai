package cmd

import (
	"log"
	"os"

	"github.com/Tech-Arch1tect/rkllmopenapi/config"
	"github.com/Tech-Arch1tect/rkllmopenapi/modelDownload"
)

func Cmd() {
	config.LoadConfig()

	if len(os.Args) != 4 {
		log.Fatal("Usage: app <username> <model_name> <model_file_name>")
	}

	// take username as first parameter
	username := os.Args[1]
	// take model name as second parameter
	modelName := os.Args[2]
	// take model file name as third parameter
	modelFileName := os.Args[3]

	modelDownload.DownloadModel(username, modelName, modelFileName)
}
