// This script assumes you have logged in through your aws credential manager and you know your bucket name and any of the csv files in the bucket
package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Presigner struct {
	PresignClient *s3.PresignClient
}

func main() {
	// Create a CPU profile file
	bucketName := "benchmark-images"
	keyName := "3gb-single.tar"
	cpuProfileFile, err := os.Create("cpu.prof")
	if err != nil {
		panic(err)
	}
	defer cpuProfileFile.Close()

	// Start CPU profiling
	if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	memProfileFile, err := os.Create("mem.prof")
	if err != nil {
		panic(err)
	}
	defer memProfileFile.Close()

	// Write memory profile to file
	if err := pprof.WriteHeapProfile(memProfileFile); err != nil {
		panic(err)
	}

	downloadImageWithS3Downloader(bucketName, keyName)
}

func downloadImageWithS3Downloader(bucketName string, key string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-1"), // Set your desired AWS region
	})
	if err != nil {
		fmt.Println("Failed to create AWS session:", err)
		return
	}

	// Create an S3 client and downloader

	downloader := s3manager.NewDownloader(sess)

	f, err := os.Create(key)
	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}

	var startTime = time.Now()
	// Download the file
	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	var totalTime = time.Since(startTime).Seconds()
	if err != nil {
		fmt.Println("Failed to download file:", err)
		return
	}

	fmt.Printf("Downloaded the given tar file in %f\n", totalTime)
	f.Close()
}

// func downloadImageWithHtcatDownloader() {
// 	sess, err := session.NewSession(&aws.Config{
// 		Region: aws.String("us-west-1"), // Set your desired AWS region
// 	})
// 	if err != nil {
// 		fmt.Println("Failed to create AWS session:", err)
// 		return
// 	}

// }

/*

	Is this the right way?
	for now, yes - It would be better if we can run htcat with a presigned url or see if that can be done easily
	How to check for memory consumption?
	How to effectively run this for htcat?
	Testing for unpack?
	Writing this into a file? would that be the same as Unpack?
	testing multilayer pull?

*/
