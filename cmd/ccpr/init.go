package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/hidetzu/ccpr/internal/config"
)

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)

	var (
		flagProfile string
		flagRegion  string
		flagForce   bool
	)

	fs.StringVar(&flagProfile, "profile", "", "AWS profile name")
	fs.StringVar(&flagRegion, "region", "", "AWS region")
	fs.BoolVar(&flagForce, "force", false, "Overwrite existing config file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	configPath, err := config.DefaultPath()
	if err != nil {
		return err
	}

	if !flagForce {
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("config file already exists: %s\n\nUse --force to overwrite:\n  ccpr init --force", configPath)
		}
	}

	profile := flagProfile
	if profile == "" {
		profile = os.Getenv("AWS_PROFILE")
	}

	region := flagRegion
	if region == "" {
		region = detectAWSRegion(profile)
	}

	cfg := &config.Config{
		Profile: profile,
		Region:  region,
	}

	if err := config.Write(configPath, cfg); err != nil {
		return err
	}

	profileDisplay := profile
	if profileDisplay == "" {
		profileDisplay = "<empty>"
	}
	regionDisplay := region
	if regionDisplay == "" {
		regionDisplay = "<empty>"
	}

	fmt.Printf("Config written to %s\n\n", configPath)
	fmt.Printf("  profile: %s\n", profileDisplay)
	fmt.Printf("  region:  %s\n", regionDisplay)
	if region == "" {
		fmt.Printf("\nNote: region is empty. Set it via:\n  - --region flag\n  - AWS_REGION env\n")
	}
	fmt.Printf("\nNext:\n  ccpr review <PR_URL>\n")
	fmt.Printf("\nIf needed, edit repoMappings in %s before running review.\n", configPath)

	return nil
}

// detectAWSRegion attempts to read the default region from AWS CLI shared config.
func detectAWSRegion(profile string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	opts := []func(*awsconfig.LoadOptions) error{}
	if profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(profile))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return ""
	}
	return awsCfg.Region
}
