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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/htcat/htcat"
)

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

	// downloadImageWithHtcatDownloader(bucketName, keyName)
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

func downloadImageWithHtcatDownloader(bucketName string, key string) {

	var startTime = time.Now()
	runTest(bucketName, key)
	var totalTime = time.Since(startTime).Seconds()
	fmt.Printf("Downloaded the given tar file in %f\n", totalTime)

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

func fetchFromHtCat(urlStr string) (io.ReadCloser, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Println("Failed to parse url", err)
	}

	hc := http.DefaultClient
	htc := htcat.New(hc, parsedURL, 1, 20)
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

func runTest(bucketName string, keyName string) {

	url := getS3PresignedUrl(bucketName, keyName)

	// fmt.Println("signed url: ", url)
	pr, err := fetchFromHtCat(url)
	if err != nil {
		fmt.Println("Failed in the test during fetchFromHtcat")
	}
	defer pr.Close()

	opFile, err := os.Create(keyName)
	if err != nil {
		fmt.Println("Failed in the test during file creation")
	}

	defer opFile.Close()

	_, err = io.Copy(opFile, pr)
	if err != nil {
		fmt.Println("Failed in the test during io Copy")
	}

	log.Println("Data fetched and written to another file")
	// rd := bufio.NewReader(stdout)
	// defer reader.Close()
	// output, err := io.ReadAll(reader)
	// if err != nil {
	// 	fmt.Println("Failed ReadAll")
	// }
	// fmt.Print(output)
}

/*

	Is io.Copy an unpack? why is it significantly slower than just s3downloader?

	Downloaded the given tar file in 25.986330 - s3Downloader
	Downloaded the given tar file in 78.803563 - htcat
	How to effectively benchmark this? Doesn't seem like it is a fair comparision.

*/
