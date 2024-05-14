// This script assumes you have logged in through your aws credential manager and you know your bucket name and any of the csv files in the bucket
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/pprof"
	"time"

	"crypto/sha256"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/htcat/htcat"
	"github.com/opencontainers/go-digest"
)

func main() {
	// Create a CPU profile file
	bucketName := "benchmark-images"
	keyName := "3gb-single.tar"
	chunkSize := 20
	// parallelism := 1
	numRuns := 5

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

	fileName := fmt.Sprintf("%s%s", "/home/ec2-user/htcat-vs-s3Downloader/", keyName)

	for j := 1; j < 4; j++ {
		for i := 0; i < numRuns; i++ {
			fmt.Printf("Running benchmark %d with parallel args %d... \n", i, j)

			os.Remove(fileName)
			// downloadImageWithS3Downloader(bucketName, keyName, j, chunkSize)
			downloadImageWithHtcatDownloader(bucketName, keyName, j, chunkSize)
			fmt.Printf("Run %d with parallel arg %d completed.\n", i, j)
			// cmd := exec.Command("sudo", "ctr", "content", "push", keyName, "test-ctr")

			// err := cmd.Run()
			// if err != nil {
			// 	log.Fatalf("Command failed: %v", err)
			// } else {
			// 	log.Println("Command executed successfully.")
			// }
		}
	}
}

func downloadImageWithS3Downloader(bucketName string, key string, parallelism int, chunkSize int) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-1"), // Set your desired AWS region
	})
	if err != nil {
		fmt.Println("Failed to create AWS session:", err)
		return
	}

	// Create an S3 client and downloader

	downloader := s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
		d.PartSize = int64(chunkSize) * 1024 * 1024 // 20MB per part
		d.Concurrency = parallelism
	})

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
	writeResult(totalTime, parallelism)
	f.Close()
}

func writeResult(totalTime float64, parallel int) {
	f, err := os.OpenFile("results.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	text := fmt.Sprintf("%s%d%s%f%s", "parallel-", parallel, ":", totalTime, "\n")
	if _, err = f.WriteString(text); err != nil {
		panic(err)
	}
}

func downloadImageWithHtcatDownloader(bucketName string, key string, parallelism int, chunkSize int) {

	var startTime = time.Now()
	runTest(bucketName, key, parallelism, chunkSize)
	var totalTime = time.Since(startTime).Seconds()
	writeResult(totalTime, parallelism)
	fmt.Printf("Time elapsed :  %f\n", totalTime)

}

func getS3PresignedUrl(bucketName string, keyName string) string {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-1")},
	)

	if err != nil {
		log.Println("Failed to create new session", err)
	}

	// Create S3 service client
	svc := s3.New(sess)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	})
	urlStr, err := req.Presign(15 * time.Minute)

	if err != nil {
		log.Println("Failed to sign request", err)
	}

	return urlStr
}

// func fetchFromHtCat(urlStr string, fileName string, parallelism int, chunkSize int) {
// 	parsedURL, err := url.Parse(urlStr)
// 	if err != nil {
// 		log.Println("Failed to parse url", err)
// 	}

// 	hc := http.DefaultClient
// 	htc := htcat.New(hc, parsedURL, parallelism, chunkSize)
// 	// fmt.Println("After htc call")
// 	opFile, err := os.Create(fileName)
// 	if err != nil {
// 		fmt.Println("Failed in the test during file creation")
// 	}

// 	defer opFile.Close()

// 	htc.WriteTo(opFile)
// }

func fetchFromHtCat(urlStr string, parallel int, chunkSize int) (io.ReadCloser, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Println("Failed to parse url", err)
	}

	hc := http.DefaultClient
	htc := htcat.New(hc, parsedURL, parallel, chunkSize)
	// fmt.Println("After htc call")
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		_, err := htc.WriteTo(pw)
		if err != nil {
			log.Println("Failed to fetch url", err)
		}
	}()
	return pr, nil
}

func runTest(bucketName string, keyName string, parallelism int, chunkSize int) {
	url := getS3PresignedUrl(bucketName, keyName)
	// fetchFromHtCat(url, keyName, parallelism, chunkSize)

	pr, err := fetchFromHtCat(url, parallelism, chunkSize)
	if err != nil {
		fmt.Println("Failed in the test during fetchFromHtcat")
	}
	defer pr.Close()

	opFile, err := os.Create(keyName)
	if err != nil {
		fmt.Println("Failed in the test during file creation")
	}

	defer opFile.Close()

	d := digest.NewDigest("sha256", sha256.New())

	mr := io.MultiReader(pr, io.TeeReader(pr, d.Algorithm().Hash()))

	_, err = io.Copy(opFile, mr)
	if err != nil {
		fmt.Println("Failed in the test during io Copy")
	}
	digestValue := d.Hex()
	fmt.Println("Digest value", digestValue)
}
