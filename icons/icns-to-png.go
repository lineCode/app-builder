package icons

import (
	"bufio"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/develar/app-builder/util"
	"github.com/disintegration/imaging"
)

type Icns2PngMapping struct {
	Id   string
	Size int
}

var icns2PngMappingList = []Icns2PngMapping{
	{"is32", 16},
	{"il32", 32},
	{"ih32", 48},
	{"icp6", 64},
	{"it32", 128},
	{"ic08", 256},
	{"ic09", 512},
}

func ConvertIcnsToPng(inFile string) (*IconListResult, error) {
	tempDir, err := util.TempDir("", ".iconset")
	if err != nil {
		return nil, err
	}

	var maxIconPath string
	var result []IconInfo

	sizeList := []int{24, 96}
	outFileTemplate := filepath.Join(tempDir, "icon_%dx%d.png")
	maxSize := 0
	if runtime.GOOS == "darwin" && os.Getenv("FORCE_ICNS2PNG") == "" {
		output, err := exec.Command("iconutil", "--convert", "iconset", "--output", tempDir, inFile).CombinedOutput()
		if err != nil {
			fmt.Println(string(output))
			return nil, err
		}

		iconFiles, err := ioutil.ReadDir(tempDir)
		if err != nil {
			return nil, err
		}

		for _, item := range icns2PngMappingList {
			fileName := fmt.Sprintf("icon_%dx%d.png", item.Size, item.Size)
			if contains(iconFiles, fileName) {
				// list sorted by size, so, last assignment is a max size
				maxIconPath = filepath.Join(tempDir, fileName)
				maxSize = item.Size
				result = append(result, IconInfo{maxIconPath, item.Size})
			} else {
				sizeList = append(sizeList, item.Size)
			}
		}
	} else {
		outputBytes, err := exec.Command("icns2png", "--extract", "--output", tempDir, inFile).CombinedOutput()
		output := string(outputBytes)
		if err != nil {
			fmt.Println(output)
			return nil, err
		}

		namePrefix := strings.TrimSuffix(filepath.Base(inFile), filepath.Ext(inFile))

		for _, item := range icns2PngMappingList {
			if strings.Contains(output, item.Id) {
				// list sorted by size, so, last assignment is a max size
				maxIconPath = filepath.Join(tempDir, fmt.Sprintf("%s_%dx%dx32.png", namePrefix, item.Size, item.Size))
				maxSize = item.Size
				result = append(result, IconInfo{maxIconPath, item.Size})
			} else {
				sizeList = append(sizeList, item.Size)
			}
		}
	}

	err = multiResizeImage(maxIconPath, outFileTemplate, &result, sizeList, maxSize)
	if err != nil {
		return nil, err
	}

	return &IconListResult{
		MaxIconPath: maxIconPath,
		MaxIconSize: maxSize,
		Icons:       result,
	}, nil
}

func contains(files []os.FileInfo, name string) bool {
	for _, fileInfo := range files {
		if fileInfo.Name() == name {
			return true
		}
	}
	return false
}

func multiResizeImage(inFile string, outFileNameFormat string, result *[]IconInfo, sizeList []int, maxSize int) (error) {
	originalImage, err := LoadImage(inFile)
	if err != nil {
		return err
	}

	var waitGroup sync.WaitGroup

	imageCount := len(sizeList)
	waitGroup.Add(imageCount)

	for i := 0; i < imageCount; i++ {
		size := sizeList[i]

		if size > maxSize {
			break
		}

		outFilePath := fmt.Sprintf(outFileNameFormat, size, size)
		*result = append(*result, IconInfo{
			File: outFilePath,
			Size: size,
		})
		go resizeImage(originalImage, size, size, outFilePath, &waitGroup)
	}

	waitGroup.Wait()
	return nil
}

func resizeImage(originalImage image.Image, w int, h int, outFileName string, waitGroup *sync.WaitGroup) error {
	defer waitGroup.Done()
	newImage := imaging.Resize(originalImage, w, h, imaging.Lanczos)
	return saveImage(newImage, outFileName)
}

func saveImage(image *image.NRGBA, outFileName string) error {
	outFile, err := os.Create(outFileName)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(outFile)
	err = png.Encode(writer, image)
	if err != nil {
		return err
	}

	flushError := writer.Flush()
	closeError := outFile.Close()
	if flushError != nil {
		return flushError
	}
	if closeError != nil {
		return closeError
	}

	return nil
}