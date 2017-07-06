// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package configure

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/urfave/cli"
)

// ConfigureCluster is the callback for ConfigureCommand (cluster).
func ConfigureCluster(context *cli.Context) {
	// get relevant fields
	region := context.String(command.RegionFlag)
	clusterProfileName := context.String(command.ConfigNameFlag)
	cluster := context.String(command.ClusterFlag)
	logrus.Warnf("Name: %s, cluster: %s, region: %s", clusterProfileName, cluster, region)

	// ecsConfig, err := createECSConfigFromCli(context)
	// if err != nil {
	// 	logrus.Error("Error initializing: ", err)
	// 	return
	// }
	// rdwr, err := config.NewReadWriter()
	// if err != nil {
	// 	logrus.Error("Error initializing: ", err)
	// 	return
	// }
	// err = saveConfig(ecsConfig, rdwr)
	// if err != nil {
	// 	logrus.Error("Error initializing: ", err)
	// }
}

// ConfigureCluster is the callback for Configure Profile subcommand.
func ConfigureProfile(context *cli.Context) {
	// get relevant fields
	secretKey := context.String(command.SecretKeyFlag)
	profileName := context.String(command.ProfileNameFlag)
	accessKey := context.String(command.AccessKeyFlag)
	logrus.Warnf("Name: %s, access: %s, secret: %s", profileName, accessKey, secretKey)
}

// ConfigureCluster is the callback for Configure Profile Default subcommand.
func ConfigureDefaultProfile(context *cli.Context) {
	// get relevant fields
	profileName := context.String(command.ProfileNameFlag)
	logrus.Warnf("Name: %s", profileName)
}

// ConfigureCluster is the callback for Configure Cluster Default subcommand.
func ConfigureDefaultCluster(context *cli.Context) {
	// get relevant fields
	clusterName := context.String(command.ConfigNameFlag)
	logrus.Warnf("Cluster Name: %s", clusterName)
}

// createECSConfigFromCli creates a new CliConfig object from the CLI context.
// It reads CLI flags to validate the ecs-cli config fields.
func createECSConfigFromCli(context *cli.Context) (*config.CliConfig, error) {
	accessKey := context.String(command.AccessKeyFlag)
	secretKey := context.String(command.SecretKeyFlag)
	region := context.String(command.RegionFlag)
	profile := context.String(command.ProfileFlag)
	cluster := context.String(command.ClusterFlag)

	if cluster == "" {
		return nil, fmt.Errorf("Missing required argument '%s'", command.ClusterFlag)
	}

	// ONLY allow for profile OR access keys to be specified
	isProfileSpecified := profile != ""
	isAccessKeySpecified := accessKey != "" || secretKey != ""
	if isProfileSpecified && isAccessKeySpecified {
		return nil, fmt.Errorf("Both AWS Access/Secret Keys and Profile were provided; only one of the two can be specified")
	}

	ecsConfig := config.NewCliConfig(cluster)
	ecsConfig.AwsProfile = profile
	ecsConfig.AwsAccessKey = accessKey
	ecsConfig.AwsSecretKey = secretKey
	ecsConfig.Region = region

	ecsConfig.ComposeProjectNamePrefix = context.String(command.ComposeProjectNamePrefixFlag)
	ecsConfig.ComposeServiceNamePrefix = context.String(command.ComposeServiceNamePrefixFlag)
	ecsConfig.CFNStackNamePrefix = context.String(command.CFNStackNamePrefixFlag)

	return ecsConfig, nil
}

// saveConfig does the actual configuration setup. This isolated method is useful for testing.
// func saveConfig(ecsConfig *config.CliConfig, rdwr config.ReadWriter) error {
//
// 	err := rdwr.Save(ecsConfig)
// 	if err != nil {
// 		return err
// 	}
// 	logrus.Infof("Saved ECS CLI configuration for cluster (%s)", ecsConfig.Cluster)
// 	return nil
// }
