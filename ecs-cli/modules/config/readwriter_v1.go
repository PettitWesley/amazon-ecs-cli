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

package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
)

const (
	clusterConfigFileName = "config.yml"
	profileConfigFileName = "profile.yml"
	configFileMode        = os.FileMode(0600)
)

// ProfileConfiguration is a simple struct for storing a single profile config
// this struct is used in the ConfigureProfile callback to save a single profile
type ProfileConfiguration struct {
	profileName  string
	awsAccessKey string
	awsSecretKey string
}

// ClusterConfiguration is a simple struct for storing a single cluster config
// this struct is used in the ConfigureCluster callback to save a single cluster
type ClusterConfiguration struct {
	clusterName string
	cluster     string
	region      string
}

// ReadWriter interface has methods to read and write ecs-cli config to and from the config file.
type ReadWriter interface {
	SaveProfile(*ProfileConfiguration) error
	SaveCluster(*ClusterConfiguration) error
	SetDefaultProfile(string) error
	SetDefaultCluster(string) error
	GetConfigs(string, string) (*CliConfig, map[interface{}]interface{}, error)
}

// YamlReadWriter implments the ReadWriter interfaces. It can be used to save and load
// ecs-cli config. Sample ecs-cli config:
// cluster: test
// aws_profile:
// region: us-west-2
// aws_access_key_id:
// aws_secret_access_key:
// compose-project-name-prefix: ecscompose-
// compose-service-name-prefix:
// cfn-stack-name-prefix: ecs-cli-
type YamlReadWriter struct {
	destination *Destination
}

// NewReadWriter creates a new Parser object.
func NewReadWriter() (*YamlReadWriter, error) {
	dest, err := newDefaultDestination()
	if err != nil {
		return nil, err
	}

	return &YamlReadWriter{destination: dest}, nil
}

// GetConfigs gets the ecs-cli config object from the config file(s).
// This function either reads the old single configuration file
// Or if the new files are present, it reads from them instead
func (rdwr *YamlReadWriter) GetConfigs(clusterConfig string, profileConfig string) (*CliConfig, map[interface{}]interface{}, error) {
	cliConfig := &CliConfig{SectionKeys: new(SectionKeys)}
	configMap := make(map[interface{}]interface{})
	// read the raw bytes of the config file
	iniPath := iniConfigPath(rdwr.destination)
	profilePath := profileConfigPath(rdwr.destination)
	clusterPath := clusterConfigPath(rdwr.destination)

	// Handle the case where the old ini config is still there
	// if ini exists and yaml does not exist, read ini
	// if both exist, read yaml
	// if neither exist, try to read yaml and then return file not found
	_, err := os.Stat(iniPath)
	_, yamlErr := os.Stat(clusterPath)
	if err == nil && yamlErr != nil { // file exists
		// old ini config
		iniReadWriter, err := NewIniReadWriter(rdwr.destination)
		if err != nil {
			return nil, nil, err
		}
		cliConfig, configMap, err = iniReadWriter.GetConfig()
		if err != nil {
			return nil, nil, err
		}

	} else {
		// If the ini file didn't exist, then we assume the yaml file exists
		// if it doesn't, then throw error
		// convert yaml to CliConfig
		clusterMap := make(map[interface{}]interface{})
		profileMap := make(map[interface{}]interface{})

		// read cluster file
		dat, err := ioutil.ReadFile(clusterPath)
		if err != nil {
			return nil, nil, err
		}

		// convert cluster yaml to a map (replaces IsKeyPresent functionality)
		if err = yaml.Unmarshal(dat, &clusterMap); err != nil {
			return nil, nil, err
		}

		// read profile file
		dat, err = ioutil.ReadFile(profilePath)
		if err != nil {
			return nil, nil, err
		}
		// convert profile yaml to a map (replaces IsKeyPresent functionality)
		if err = yaml.Unmarshal(dat, &profileMap); err != nil {
			return nil, nil, err
		}

		logrus.Warnf("c: %s, p: %s", clusterConfig, profileConfig)
		processProfileMap(profileConfig, profileMap, configMap, cliConfig)
		processClusterMap(clusterConfig, clusterMap, configMap, cliConfig)

	}
	return cliConfig, configMap, nil
}

func processProfileMap(profileKey string, profileMap map[interface{}]interface{}, configMap map[interface{}]interface{}, cliConfig *CliConfig) error {
	if profileKey == "" {
		var ok bool
		profileKey, ok = profileMap["default"].(string)
		if !ok {
			return errors.New("Format issue with profile config file; expected key not found.")
		}
	}
	profile, ok := profileMap["ecs_profiles"].(map[interface{}]interface{})[profileKey].(map[interface{}]interface{})
	if !ok {
		return errors.New("Format issue with profile config file; expected key not found.")
	}

	configMap[awsAccessKey] = profile[awsAccessKey]
	configMap[awsSecretKey] = profile[awsSecretKey]
	cliConfig.AwsAccessKey, ok = profile[awsAccessKey].(string)
	if !ok {
		return errors.New("Format issue with profile config file; expected key not found.")
	}
	cliConfig.AwsSecretKey, ok = profile[awsSecretKey].(string)
	if !ok {
		return errors.New("Format issue with profile config file; expected key not found.")
	}

	return nil

}

func processClusterMap(clusterConfigKey string, clusterMap map[interface{}]interface{}, configMap map[interface{}]interface{}, cliConfig *CliConfig) error {
	if clusterConfigKey == "" {
		var ok bool
		clusterConfigKey, ok = clusterMap["default"].(string)
		if !ok {
			return errors.New("Format issue with cluster config file; expected key not found.")
		}
	}
	cluster, ok := clusterMap["clusters"].(map[interface{}]interface{})[clusterConfigKey].(map[interface{}]interface{})
	if !ok {
		return errors.New("Format issue with cluster config file; expected key not found.")
	}

	configMap[clusterKey] = cluster[clusterKey]
	logrus.Warnf("Cluster from file: %s", cluster[clusterKey])
	configMap[regionKey] = cluster[regionKey]
	cliConfig.Cluster, ok = cluster[clusterKey].(string)
	if !ok {
		return errors.New("Format issue with cluster config file; expected key not found.")
	}
	cliConfig.Region, ok = cluster[regionKey].(string)
	if !ok {
		return errors.New("Format issue with cluster config file; expected key not found.")
	}

	// Prefixes
	// ComposeProjectNamePrefix
	if _, ok := cluster[composeProjectNamePrefixKey]; ok {
		configMap[composeProjectNamePrefixKey] = cluster[composeProjectNamePrefixKey]
		cliConfig.ComposeProjectNamePrefix, ok = cluster[composeProjectNamePrefixKey].(string)
		if !ok {
			return errors.New("Format issue with cluster config file; expected key not found.")
		}
	}
	// ComposeServiceNamePrefix
	if _, ok := cluster[composeServiceNamePrefixKey]; ok {
		configMap[composeServiceNamePrefixKey] = cluster[composeServiceNamePrefixKey]
		cliConfig.ComposeServiceNamePrefix, ok = cluster[composeServiceNamePrefixKey].(string)
		if !ok {
			return errors.New("Format issue with cluster config file; expected key not found.")
		}
	}
	// CFNStackNamePrefix
	if _, ok := cluster[cfnStackNamePrefixKey]; ok {
		configMap[cfnStackNamePrefixKey] = cluster[cfnStackNamePrefixKey]
		cliConfig.CFNStackNamePrefix, ok = cluster[cfnStackNamePrefixKey].(string)
		if !ok {
			return errors.New("Format issue with profile cluster file; expected key not found.")
		}
	}

	return nil

}

func (rdwr *YamlReadWriter) SetDefaultProfile(profile string) error {
	profileMap := make(map[interface{}]interface{})
	profilePath := profileConfigPath(rdwr.destination)

	// read profile file
	dat, err := ioutil.ReadFile(profilePath)
	if err != nil {
		return err
	}
	// convert profile yaml to a map
	if err = yaml.Unmarshal(dat, &profileMap); err != nil {
		return err
	}

	profileMap["default"] = profile

	// we must save the entire new config map to the file
	rdwr.saveToFile(profilePath, profileMap)

	return nil
}

func (rdwr *YamlReadWriter) SetDefaultCluster(cluster string) error {
	clusterMap := make(map[interface{}]interface{})
	clusterPath := clusterConfigPath(rdwr.destination)

	// read profile file
	dat, err := ioutil.ReadFile(clusterPath)
	if err != nil {
		return err
	}
	// convert profile yaml to a map
	if err = yaml.Unmarshal(dat, &clusterMap); err != nil {
		return err
	}

	clusterMap["default"] = cluster

	// we must save the entire new config map to the file
	rdwr.saveToFile(clusterPath, clusterMap)

	return nil
}

func (rdwr *YamlReadWriter) SaveProfile(profile *ProfileConfiguration) error {
	profileMap := make(map[interface{}]interface{})
	profilePath := profileConfigPath(rdwr.destination)

	// read profile file
	// we must read the profile to see the existing config
	// a new config with the same name will lead to replacement
	// if this is the first config to be defined, then we make it default
	dat, err := ioutil.ReadFile(profilePath)
	if err != nil {
		return err
	}
	// convert profile yaml to a map
	if err = yaml.Unmarshal(dat, &profileMap); err != nil {
		return err
	}

	profiles, ok := profileMap["ecs_profiles"].(map[interface{}]interface{})
	if !ok {
		return errors.New("Format issue with profile config file; expected key not found.")
	}

	if len(profiles) == 0 { // this is the first one to be defined; make default
		profileMap["default"] = profile.profileName
	}

	newProfile := make(map[interface{}]interface{})
	newProfile[awsAccessKey] = profile.awsAccessKey
	newProfile[awsSecretKey] = profile.awsSecretKey

	profiles[profile.profileName] = newProfile

	// we must save the entire new config map to the file
	rdwr.saveToFile(profilePath, profileMap)

	return nil

}

func (rdwr *YamlReadWriter) SaveCluster(cluster *ClusterConfiguration) error {
	clusterMap := make(map[interface{}]interface{})
	clusterPath := clusterConfigPath(rdwr.destination)

	// read profile file
	// we must read the profile to see the existing config
	// a new config with the same name will lead to replacement
	// if this is the first config to be defined, then we make it default
	dat, err := ioutil.ReadFile(clusterPath)
	if err != nil {
		return err
	}
	// convert profile yaml to a map
	if err = yaml.Unmarshal(dat, &clusterMap); err != nil {
		return err
	}

	clusters, ok := clusterMap["clusters"].(map[interface{}]interface{})
	if !ok {
		return errors.New("Format issue with cluster config file; expected key not found.")
	}

	if len(clusters) == 0 { // this is the first one to be defined; make default
		clusterMap["default"] = cluster.clusterName
	}

	newCluster := make(map[interface{}]interface{})
	newCluster[clusterKey] = cluster.cluster
	newCluster[regionKey] = cluster.region

	clusters[cluster.clusterName] = newCluster

	// we must save the entire new config map to the file
	rdwr.saveToFile(clusterPath, clusterMap)

	return nil

}

func (rdwr *YamlReadWriter) saveToFile(path string, config map[interface{}]interface{}) error {
	destMode := rdwr.destination.Mode
	err := os.MkdirAll(rdwr.destination.Path, *destMode)
	if err != nil {
		return err
	}

	// Warn the user if in path also exists
	iniPath := iniConfigPath(rdwr.destination)
	_, iniErr := os.Stat(iniPath)
	if iniErr == nil {
		logrus.Warnf("Writing yaml formatted config to %s/.ecs/.\nIni formatted config still exists in %s/.ecs/%s.", os.Getenv("HOME"), os.Getenv("HOME"), iniConfigFileName)
	}

	// If config file exists, set permissions first, because we may be writing creds.
	if _, err := os.Stat(path); err == nil {
		if err = os.Chmod(path, configFileMode); err != nil {
			logrus.Errorf("Unable to chmod %s to mode %s", path, configFileMode)
			return err
		}
	}

	data, err := yaml.Marshal(config)
	err = ioutil.WriteFile(path, data, configFileMode.Perm())
	if err != nil {
		logrus.Errorf("Unable to write config to %s", path)
		return err
	}

	return nil
}

func profileConfigPath(dest *Destination) string {
	return filepath.Join(dest.Path, profileConfigFileName)
}

func clusterConfigPath(dest *Destination) string {
	return filepath.Join(dest.Path, clusterConfigFileName)
}

func iniConfigPath(dest *Destination) string {
	return filepath.Join(dest.Path, iniConfigFileName)
}
