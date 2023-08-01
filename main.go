package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type EC2DescribeInstancesAPI interface {
	DescribeInstances(ctx context.Context,
		params *ec2.DescribeInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

func GetInstances(c context.Context, api EC2DescribeInstancesAPI, input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return api.DescribeInstances(c, input)
}

func buildAttributeFilterListV2(m map[string]string) []ec2types.Filter {
	var filters []ec2types.Filter

	// sort the filters by name to make the output deterministic
	var names []string
	for k := range m {
		names = append(names, k)
	}

	sort.Strings(names)

	for _, name := range names {
		value := m[name]
		if value == "" {
			continue
		}

		filters = append(filters, newFilterV2(name, []string{value}))
	}

	return filters
}

func newFilterV2(name string, values []string) ec2types.Filter {
	return ec2types.Filter{
		Name:   aws.String(name),
		Values: values,
	}
}

func DescribeInstancesCmd(awsRegion *string, amiAge *int, amiPattern *string) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(*awsRegion))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	currentTime := time.Now()
	pastTime := currentTime.AddDate(0, 0, -*amiAge)

	client := ec2.NewFromConfig(cfg)

	input := &ec2.DescribeInstancesInput{Filters: buildAttributeFilterListV2(map[string]string{"launch-time": pastTime.Format(time.RFC3339)})}

	result, err := GetInstances(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error retrieving information about your Amazon EC2 instances:")
		fmt.Println(err)
		return
	}

	for _, r := range result.Reservations {
		fmt.Println("Reservation ID: " + *r.ReservationId)
		fmt.Println("Instance IDs:")
		for _, i := range r.Instances {
			fmt.Println("   " + *i.InstanceId)
		}

		fmt.Println("")
	}
}

func main() {

	var amiPattern string // The search string for the AMIs you want to find. Can be a glob.
	var awsRegion string  // The region the AMIs are stored in
	var amiAge int        // Find AMIs older than this value in days
	flag.StringVar(&amiPattern, "pattern", "*", "Pattern to match AMI names")
	flag.StringVar(&awsRegion, "region", "us-west-2", "AWS region to use")
	flag.IntVar(&amiAge, "age", 90, "Age of AMI in days")
	flag.Parse()

	result, err := DescribeInstancesCmd()
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

	amiParams := &ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("name"),
				Values: []*string{aws.String(amiPattern)},
			},
		},
	}
	amiResult, err := &ec2.DescribeImages(amiParams)
	if err != nil {
		fmt.Println("Error describing AMIs: ", err)
		os.Exit(1)
	}

	unusedAmis := make([]string, 0)
	for _, image := range amiResult.Images {
		if _, used := usedAmiIds[*image.ImageId]; !used && strings.Contains(*image.Name, amiPattern) {
			unusedAmis = append(unusedAmis, *image.ImageId)
		}
	}

	fmt.Println("Unused AMIs:")
	for _, amiId := range unusedAmis {
		fmt.Println(amiId)
	}
}
