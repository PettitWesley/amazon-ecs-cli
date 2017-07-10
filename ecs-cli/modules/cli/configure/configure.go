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

	clusterConfig := &config.ClusterConfiguration{ClusterProfileName: clusterProfileName, Cluster: cluster, Region: region}

	// modify the profile config file
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error initializing: ", err)
		return
	}
	if err = rdwr.SaveCluster(clusterConfig); err != nil {
		logrus.Error("Error initializing: ", err)
	}
}

// ConfigureCluster is the callback for Configure Profile subcommand.
func ConfigureProfile(context *cli.Context) {
	// get relevant fields
	secretKey := context.String(command.SecretKeyFlag)
	profileName := context.String(command.ProfileNameFlag)
	accessKey := context.String(command.AccessKeyFlag)

	profile := &config.ProfileConfiguration{ProfileName: profileName, AwsAccessKey: accessKey, AwsSecretKey: secretKey}

	// modify the profile config file
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error initializing: ", err)
		return
	}
	if err = rdwr.SaveProfile(profile); err != nil {
		logrus.Error("Error initializing: ", err)
	}

}

// ConfigureCluster is the callback for Configure Profile Default subcommand.
func ConfigureDefaultProfile(context *cli.Context) {
	// get relevant fields
	profileName := context.String(command.ProfileNameFlag)

	// modify the profile config file
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error initializing: ", err)
		return
	}
	if err = rdwr.SetDefaultProfile(profileName); err != nil {
		logrus.Error("Error initializing: ", err)
	}

}

// ConfigureCluster is the callback for Configure Cluster Default subcommand.
func ConfigureDefaultCluster(context *cli.Context) {
	// get relevant fields
	clusterName := context.String(command.ConfigNameFlag)

	// modify the profile config file
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error initializing: ", err)
		return
	}
	if err = rdwr.SetDefaultCluster(clusterName); err != nil {
		logrus.Error("Error initializing: ", err)
	}
}
