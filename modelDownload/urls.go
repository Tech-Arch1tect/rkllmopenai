package modelDownload

import (
	"fmt"
	"log"
	"os"

	"github.com/Tech-Arch1tect/rkllmopenapi/config"
)

var (
	HuggingFaceUrl = "https://huggingface.co/%s/%s/resolve/main/%s?download=true"
)

func DownloadModel(Username, ModelName, ModelFileName string) {
	tokenizerUrl := fmt.Sprintf(HuggingFaceUrl, Username, ModelName, "tokenizer.json")
	tokenizerConfigUrl := fmt.Sprintf(HuggingFaceUrl, Username, ModelName, "tokenizer_config.json")
	modelUrl := fmt.Sprintf(HuggingFaceUrl, Username, ModelName, ModelFileName)

	storagePath := config.C.StoragePath

	// check if any files exist in the storage path, if not create the directory, if yes error
	if _, err := os.Stat(storagePath + "/" + Username + "/" + ModelName); os.IsNotExist(err) {
		os.MkdirAll(storagePath+"/"+Username+"/"+ModelName, 0755)
	} else {
		log.Fatal("Model files already exist in", storagePath+"/"+Username+"/"+ModelName)
	}

	// download tokenizer, tokenizer_config, model
	err := downloadFile(tokenizerUrl, storagePath+"/"+Username+"/"+ModelName+"/tokenizer.json")
	if err != nil {
		log.Fatal("Error downloading tokenizer.json:", err)
	}
	err = downloadFile(tokenizerConfigUrl, storagePath+"/"+Username+"/"+ModelName+"/tokenizer_config.json")
	if err != nil {
		log.Fatal("Error downloading tokenizer_config.json:", err)
	}
	err = downloadFile(modelUrl, storagePath+"/"+Username+"/"+ModelName+"/"+ModelFileName)
	if err != nil {
		log.Fatal("Error downloading model:", err)
	}

	fmt.Println("Downloaded model files to", storagePath+"/"+Username+"/"+ModelName)
}
