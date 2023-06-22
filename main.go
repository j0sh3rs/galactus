package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func main() {
	// Check for AWS credentials
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		fmt.Println("ERROR: AWS credentials are not set in environment variables")
		os.Exit(1)
	}

	// Parse command-line arguments for AMI name pattern and AWS region
	var amiPattern string
	var awsRegion string
	flag.StringVar(&amiPattern, "pattern", "*", "Pattern to match AMI names")
	flag.StringVar(&awsRegion, "region", "us-west-2", "AWS region to use") // Default to us-west-2
	flag.Parse()

	// Establish AWS session
	sess, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		fmt.Println("Error creating AWS session: ", err)
		os.Exit(1)
	}

	// Create new EC2 service
	ec2Svc := ec2.NewFromConfig(sess)

	// Call DescribeInstances to get all instances launched in the last 90 days
	currentTime := time.Now()
	pastTime := currentTime.AddDate(0, 0, -90)
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filters[{
			{
				Name:   "launch-time",
				Values: pastTime.Format(time.RFC3339),
			},
		},
	}]
	result, err := ec2Svc.DescribeInstances(params)
	if err != nil {
		fmt.Println("Error describing instances: ", err)
		os.Exit(1)
	}

	usedAmiIds := make(map[string]bool)
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			usedAmiIds[*instance.ImageId] = true
		}
	}

	// Call DescribeImages to get all AMIs
	amiParams := &ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name:   "name",
				Values: &amiPattern,
			},
		},
	}
	amiResult, err := ec2Svc.DescribeImages(amiParams)
	if err != nil {
		fmt.Println("Error describing AMIs: ", err)
		os.Exit(1)
	}

	// Check if AMIs have been used
	unusedAmis := make([]string, 0)
	for _, image := range amiResult.Images {
		if _, used := usedAmiIds[*image.ImageId]; !used && strings.Contains(*image.Name, amiPattern) {
			unusedAmis = append(unusedAmis, *image.ImageId)
		}
	}

	// Output unused AMIs
	fmt.Println("Unused AMIs:")
	for _, amiId := range unusedAmis {
		fmt.Println(amiId)
	}
}
