package imagestore

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/IDarar/hub/pkg/logger"
)

type DiskImageStore struct {
	ImageFolder string
}

func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore{
		ImageFolder: imageFolder,
	}
}

func (s *DiskImageStore) Save(ext string, imageData *bytes.Buffer) (string, error) {

	imageID := RandomString()

	imagePath := fmt.Sprintf("%s/%s.%s", s.ImageFolder, imageID, ext)

	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %w", err)
	}

	n, err := imageData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}

	logger.Info("bytes written: ", n)

	return imageID + "." + ext, nil
}

func RandomString() string {
	rand.Seed(int64(time.Now().UnixNano()))

	charSet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstxyz0123456789"
	var output strings.Builder
	length := 18
	for i := 0; i < length; i++ {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		output.WriteString(string(randomChar))
	}
	return output.String()
}
