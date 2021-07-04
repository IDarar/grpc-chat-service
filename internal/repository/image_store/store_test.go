package imagestore

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/IDarar/hub/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestSave(t *testing.T) {
	s := DiskImageStore{ImageFolder: "./test_images"}

	testCases := []struct {
		name      string
		imagePath string
		ext       string
	}{
		{
			name:      "successful save image",
			imagePath: "./tmp/1.png",
			ext:       "png",
		},
		{

			name:      "successful save image",
			imagePath: "./tmp/3.png",
			ext:       "png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//prepare bytes from test image
			imageBytes, err := ioutil.ReadFile(tc.imagePath)
			require.NoError(t, err)

			imageData := &bytes.Buffer{}

			imageData.Write(imageBytes)

			logger.Info("len of buffer before saving", imageData.Len())

			id, err := s.Save(tc.ext, imageData)
			require.NoError(t, err)

			logger.Info(id)

			logger.Info("len of buffer after saving", imageData.Len())

			savedImagePath := fmt.Sprintf("%s/%s", s.ImageFolder, id)

			logger.Info(savedImagePath)

			require.FileExists(t, savedImagePath)
			require.NoError(t, os.Remove(savedImagePath))
		})
	}
}
