package main

import (
	"fmt"

	"github.com/brendan-ward/rastertiler/gdal"
)

func main() {
	// fmt.Printf("GDAL version: %v\n", gdal.GDALVersion)

	d, err := gdal.Open("/tmp/blueprint2021.tif")
	if err != nil {
		panic(err)
	}
	defer d.Close()

	fmt.Println(d)
}
