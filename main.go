package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"path/filepath"
	"strings"

	"gocv.io/x/gocv"
)

func main() {
	imageFilename := "image.jpg"

	img := gocv.IMRead(imageFilename, gocv.IMReadUnchanged)
	defer closeImage(&img)

	imgGray := gocv.NewMat()
	defer closeImage(&imgGray)
	gocv.CvtColor(img, &imgGray, gocv.ColorBGRToGray)
	gocv.IMWrite(fileNameExtend(imageFilename, "gray"), imgGray)

	imgBlur := gocv.NewMat()
	defer closeImage(&imgBlur)
	gocv.GaussianBlur(imgGray, &imgBlur, image.Point{X: 9, Y: 9}, 0, 0, gocv.BorderDefault)
	gocv.IMWrite(fileNameExtend(imageFilename, "blur"), imgBlur)

	imgThresh := gocv.NewMat()
	defer closeImage(&imgThresh)
	gocv.Threshold(imgBlur, &imgThresh, 0, 255, gocv.ThresholdBinaryInv+gocv.ThresholdOtsu)
	gocv.IMWrite(fileNameExtend(imageFilename, "thresh"), imgThresh)

	imgDilate := gocv.NewMat()
	defer closeImage(&imgDilate)
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 30, Y: 5})
	gocv.Dilate(imgThresh, &imgDilate, kernel)
	gocv.IMWrite(fileNameExtend(imageFilename, "dilate"), imgDilate)

	contours := gocv.FindContours(imgDilate, gocv.RetrievalList, gocv.ChainApproxSimple)
	fmt.Printf("Contours.size: %d\n", contours.Size())

	imgContours := img.Clone()
	defer closeImage(&imgContours)
	boxColor := color.RGBA{R: 255, A: 1}
	largestContour := float64(0)
	largestContourIdx := 0

	for i := 0; i < contours.Size(); i++ {
		gocv.DrawContours(&imgContours, contours, i, boxColor, 2)

		rect := gocv.BoundingRect(contours.At(i))
		gocv.Rectangle(&imgContours, rect, color.RGBA{B: 255}, 2)

		area := gocv.ContourArea(contours.At(i))
		if area > largestContour {
			largestContour = area
			largestContourIdx = i
		}
		fmt.Printf("Contour[%d] area: %f\n", i, area)
	}

	fmt.Printf("Largest Contour area is [%d]: %f\n", largestContourIdx, gocv.ContourArea(contours.At(largestContourIdx)))

	gocv.IMWrite(fileNameExtend(imageFilename, "contours"), imgContours)

	minAreaRect := gocv.MinAreaRect(contours.At(largestContourIdx))

	fmt.Print("MinAreaRect Points:\n")
	for _, v := range minAreaRect.Points {
		fmt.Printf(" x: %d  y: %d\n", v.X, v.Y)
	}

	angle := minAreaRect.Angle
	if angle < -45 {
		angle = 90 + angle
	}
	// angle *= -1
	fmt.Printf("Angle: %f\n", angle)

	imgRotated := img.Clone()
	defer closeImage(&imgRotated)
	center := image.Point{X: img.Cols() / 2, Y: img.Rows() / 2}
	rotationMatrix := gocv.GetRotationMatrix2D(center, angle, 1)
	gocv.WarpAffine(img, &imgRotated, rotationMatrix, image.Point{})

	gocv.IMWrite(fileNameExtend(imageFilename, "rotated"), imgRotated)
}

// fileNameExtend adds ext string to end of the base of the fileName.
// Example: file.jpg -> file_ext.jpg
func fileNameExtend(fileName, ext string) string {
	e := filepath.Ext(fileName)
	b := strings.TrimSuffix(fileName, e)
	return b + "_" + ext + e
}

func closeImage(img *gocv.Mat) {
	err := img.Close()
	if err != nil {
		log.Fatal(err)
	}
}
