module benchmarks/htcat-vs-s3Downloader

go 1.20

require github.com/aws/aws-sdk-go v1.51.21

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	golang.org/x/net v0.24.0 // indirect
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.5 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/htcat/htcat v1.0.2
)

require (
	github.com/aws/aws-sdk-go-v2 v1.26.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.53.1
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)

replace github.com/htcat/htcat v1.0.2 => /home/ec2-user/htcat
