package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hidetzu/ccpr/internal/config"
)

type checkResult struct {
	ok      bool
	message string
	fix     string
}

func runDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)

	var (
		flagConfig  string
		flagProfile string
		flagRegion  string
	)

	fs.StringVar(&flagConfig, "config", "", "Path to configuration file")
	fs.StringVar(&flagProfile, "profile", "", "AWS profile name")
	fs.StringVar(&flagRegion, "region", "", "AWS region")

	if err := fs.Parse(args); err != nil {
		return err
	}

	var results []checkResult

	// 1. Config file check
	cfg, cfgResult := checkConfig(flagConfig)
	results = append(results, cfgResult)

	// 2. AWS credentials check
	profile := ""
	region := flagRegion
	if cfg != nil {
		profile = cfg.ResolveProfile(flagProfile)
		if region == "" {
			region = cfg.ResolveRegion("")
		}
	} else if flagProfile != "" {
		profile = flagProfile
	}
	results = append(results, checkAWSCredentials(profile, region))

	// 3. Repo mappings check (sorted for stable output)
	if cfg != nil && len(cfg.RepoMappings) > 0 {
		names := make([]string, 0, len(cfg.RepoMappings))
		for name := range cfg.RepoMappings {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			results = append(results, checkRepoMapping(name, cfg.RepoMappings[name]))
		}
	}

	// Print results
	passed := 0
	total := len(results)
	for _, r := range results {
		if r.ok {
			fmt.Printf("✔ %s\n", r.message)
			passed++
		} else {
			fmt.Printf("✖ %s\n", r.message)
			if r.fix != "" {
				fmt.Printf("  → %s\n", r.fix)
			}
		}
	}

	fmt.Printf("\n%d/%d checks passed\n", passed, total)

	if passed < total {
		return fmt.Errorf("%d check(s) failed", total-passed)
	}
	return nil
}

func checkConfig(flagConfig string) (*config.Config, checkResult) {
	cfg, resolvedPath, err := config.Load(flagConfig)
	if err != nil {
		displayPath := resolvedPath
		if displayPath == "" {
			if p, e := config.DefaultPath(); e == nil {
				displayPath = p
			}
		}
		return nil, checkResult{
			ok:      false,
			message: fmt.Sprintf("Config file: %s", displayPath),
			fix:     "Run: ccpr init",
		}
	}

	return cfg, checkResult{
		ok:      true,
		message: fmt.Sprintf("Config file: %s", resolvedPath),
	}
}

func checkAWSCredentials(profile, region string) checkResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := []func(*awsconfig.LoadOptions) error{}
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}
	if profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(profile))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return checkResult{
			ok:      false,
			message: "AWS credentials",
			fix:     "Configure AWS credentials: aws configure",
		}
	}

	stsClient := sts.NewFromConfig(awsCfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		hint := "Configure AWS credentials: aws configure"
		if profile != "" {
			hint = fmt.Sprintf("Check profile %q: aws configure --profile %s", profile, profile)
		}
		return checkResult{
			ok:      false,
			message: "AWS credentials",
			fix:     hint,
		}
	}

	arn := ""
	if identity.Arn != nil {
		arn = *identity.Arn
	}
	return checkResult{
		ok:      true,
		message: fmt.Sprintf("AWS credentials: %s", arn),
	}
}

func checkRepoMapping(name, path string) checkResult {
	info, err := os.Stat(path)
	if err != nil {
		return checkResult{
			ok:      false,
			message: fmt.Sprintf("Repo mapping: %s → %s (path not found)", name, path),
			fix:     "Check repoMappings in config file",
		}
	}
	if !info.IsDir() {
		return checkResult{
			ok:      false,
			message: fmt.Sprintf("Repo mapping: %s → %s (not a directory)", name, path),
			fix:     "Check repoMappings in config file",
		}
	}

	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return checkResult{
			ok:      false,
			message: fmt.Sprintf("Repo mapping: %s → %s (not a git repository)", name, path),
			fix:     fmt.Sprintf("Run: git -C %s init", path),
		}
	}

	return checkResult{
		ok:      true,
		message: fmt.Sprintf("Repo mapping: %s → %s (git OK)", name, path),
	}
}
