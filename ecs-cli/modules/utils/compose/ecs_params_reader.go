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

package utils

// ECS Params Reader is used to parse the ecs-params.yml file and marshal the data into the ECSParams struct

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/cli/cli/compose/types"
	libYaml "github.com/docker/libcompose/yaml"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// ECSParams contains the information parsed from the ecs-params.yml file
type ECSParams struct {
	Version        string
	TaskDefinition EcsTaskDef `yaml:"task_definition"`
	RunParams      RunParams  `yaml:"run_params"`
}

// EcsTaskDef corresponds to fields in an ECS TaskDefinition
type EcsTaskDef struct {
	NetworkMode          string        `yaml:"ecs_network_mode"`
	TaskRoleArn          string        `yaml:"task_role_arn"`
	ContainerDefinitions ContainerDefs `yaml:"services"`
	ExecutionRole        string        `yaml:"task_execution_role"`
	TaskSize             TaskSize      `yaml:"task_size"` // Needed to run FARGATE tasks
}

// ContainerDefs is a map of ContainerDefs within a task definition
type ContainerDefs map[string]ContainerDef

// ContainerDef holds fields for an ECS Container Definition that are not supplied by docker-compose
type ContainerDef struct {
	Essential bool `yaml:"essential"`
	// resource field yaml names correspond to equivalent docker-compose field
	Cpu               int64                  `yaml:"cpu_shares"`
	Memory            libYaml.MemStringorInt `yaml:"mem_limit"`
	MemoryReservation libYaml.MemStringorInt `yaml:"mem_reservation"`
	HealthCheck       HealthCheck            `yaml:"healthcheck"`
}

// HealthCheck holds the ECS container health check
// https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_HealthCheck.html
type HealthCheck struct {
	ecs.HealthCheck
}

// healthCheckFormat is used to unmarshal the different healthcheck formats supported by ECS Params
type healthCheckFormat struct {
	Test        libYaml.Stringorslice
	Command     libYaml.Stringorslice
	Timeout     string `yaml:"timeout,omitempty"`
	Interval    string `yaml:"interval,omitempty"`
	Retries     int64  `yaml:"retries,omitempty"`
	StartPeriod string `yaml:"start_period,omitempty"`
}

type healthCheckComposeFormat struct {
	Test        []string       `yaml:"test,omitempty"`
	Timeout     *time.Duration `yaml:"timeout,omitempty"`
	Interval    *time.Duration `yaml:"interval,omitempty"`
	Retries     *uint64        `yaml:"retries,omitempty"`
	StartPeriod *time.Duration `yaml:"start_period,omitempty"`
}

type healthCheckECSFormat struct {
	Command     []*string `yaml:"command,omitempty"`
	Timeout     *int64    `yaml:"timeout,omitempty"`
	Interval    *int64    `yaml:"interval,omitempty"`
	Retries     *int64    `yaml:"retries,omitempty"`
	StartPeriod *int64    `yaml:"start_period,omitempty"`
}

type healthCheckWithTestAsString struct {
	Test string `yaml:"test,omitempty"`
}

// TaskSize holds Cpu and Memory values needed for Fargate tasks
// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-cpu-memory-error.html
type TaskSize struct {
	Cpu    string `yaml:"cpu_limit"`
	Memory string `yaml:"mem_limit"`
}

// RunParams specifies non-TaskDefinition specific parameters
type RunParams struct {
	NetworkConfiguration NetworkConfiguration `yaml:"network_configuration"`
}

// NetworkConfiguration specifies the network config for the task definition.
// Supports values 'awsvpc' (required for Fargate), 'bridge', 'host' or 'none'
type NetworkConfiguration struct {
	AwsVpcConfiguration AwsVpcConfiguration `yaml:"awsvpc_configuration"`
}

// AwsVpcConfiguration specifies the networking resources available to
// tasks running in 'awsvpc' networking mode
type AwsVpcConfiguration struct {
	Subnets        []string       `yaml:"subnets"`
	SecurityGroups []string       `yaml:"security_groups"`
	AssignPublicIp AssignPublicIp `yaml:"assign_public_ip"` // Needed to run FARGATE tasks
}

type AssignPublicIp string

const (
	Enabled  AssignPublicIp = "ENABLED"
	Disabled AssignPublicIp = "DISABLED"
)

func (cd *ContainerDef) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawContainerDef ContainerDef
	raw := rawContainerDef{Essential: true} //  If essential is not specified, we want it to be true
	if err := unmarshal(&raw); err != nil {
		return err
	}

	*cd = ContainerDef(raw)
	return nil
}

// HealthCheck.UnmarshalYAML is a custom unmarshaler for healthcheck that parses
// both the docker compose and ecs inspired syntaxes
func (h *HealthCheck) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// All of the different health check formats are mutually exclusive in each of their fields
	// This makes parsing simple.
	healthCheck := HealthCheck{}

	// set default value for retries
	rawHealthCheck := healthCheckFormat{
		Retries: 3,
	}
	if err := unmarshal(&rawHealthCheck); err != nil {
		return err
	}

	if len(rawHealthCheck.Command) > 0 && len(rawHealthCheck.Test) > 0 {
		return fmt.Errorf("healthcheck.test and healthcheck.command can not both be specified")
	}

	if len(rawHealthCheck.Command) > 0 {
		parseHealthCheckCommand(rawHealthCheck.Command, &healthCheck)
	}

	if len(rawHealthCheck.Test) > 0 {
		parseHealthCheckCommand(rawHealthCheck.Test, &healthCheck)
	}

	healthCheck.SetRetries(rawHealthCheck.Retries)

	if timeout, err := parseHealthCheckTimeField(rawHealthCheck.Timeout); err == nil {
		healthCheck.Timeout = timeout
	} else {
		return err
	}

	if startPeriod, err := parseHealthCheckTimeField(rawHealthCheck.StartPeriod); err == nil {
		healthCheck.StartPeriod = startPeriod
	} else {
		return err
	}

	if interval, err := parseHealthCheckTimeField(rawHealthCheck.Interval); err == nil {
		healthCheck.Interval = interval
	} else {
		return err
	}

	*h = healthCheck

	return nil
}

// parses the command/test field for healthcheck
func parseHealthCheckCommand(command []string, healthCheck *HealthCheck) {
	if len(command) == 1 {
		// command/test was specified as a string which wraps it in /bin/sh
		healthCheck.SetCommand(aws.StringSlice([]string{
			"CMD-SHELL",
			command[0],
		}))
	} else {
		healthCheck.SetCommand(aws.StringSlice(command))
	}
}

// parses a health check time string which could be a duration or an integer
func parseHealthCheckTimeField(field string) (*int64, error) {
	if field != "" {
		duration, err := time.ParseDuration(field)
		if err == nil {
			return adapter.ConvertToTimeInSeconds(&duration), nil
		} else if val, err := strconv.ParseInt(field, 10, 64); err == nil {
			return &val, nil
		} else {
			return nil, fmt.Errorf("Could not parse %s either as an integer or a duration (ex: 1m30s)", field)
		}
	}

	return nil, nil
}

// func (h *HealthCheck) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	// All of the different health check formats are mutually exclusive in each of their fields
// 	// This makes parsing simple.
// 	healthCheck := HealthCheck{}
//
// 	var healthCheckCompose healthCheckComposeFormat
// 	unmarshal(&healthCheckCompose)
// 	healthCheckCompose.toHealthCheck(&healthCheck)
//
// 	var healthCheckECS healthCheckECSFormat
// 	unmarshal(&healthCheckECS)
// 	healthCheckECS.toHealthCheck(&healthCheck)
//
// 	var testCommandAltFormat healthCheckWithTestAsString
// 	unmarshal(&testCommandAltFormat)
// 	testCommandAltFormat.toHealthCheck(&healthCheck)
//
// 	*h = healthCheck
//
// 	return nil
// }

func (h *healthCheckECSFormat) toHealthCheck(healthCheck *HealthCheck) {
	if len(h.Command) > 0 {
		healthCheck.Command = h.Command
	}
	if h.Interval != nil && *h.Interval != 0 {
		healthCheck.Interval = h.Interval
	}
	if h.Retries != nil && *h.Retries != 0 {
		healthCheck.Retries = h.Retries
	}
	if h.StartPeriod != nil && *h.StartPeriod != 0 {
		healthCheck.StartPeriod = h.StartPeriod
	}
	if h.Timeout != nil && *h.Timeout != 0 {
		healthCheck.Timeout = h.Timeout
	}
}

func (h *healthCheckComposeFormat) toHealthCheck(healthCheck *HealthCheck) {
	healthCheck.HealthCheck = *adapter.ConvertToHealthCheck(&types.HealthCheckConfig{
		Test:        h.Test,
		Timeout:     h.Timeout,
		Interval:    h.Interval,
		Retries:     h.Retries,
		StartPeriod: h.StartPeriod,
	})
}

func (h *healthCheckWithTestAsString) toHealthCheck(healthCheck *HealthCheck) {
	// specifying health check command as string wraps it in /bin/sh
	if h.Test != "" {
		command := []string{"CMD-SHELL", h.Test}
		healthCheck.SetCommand(aws.StringSlice(command))
	}
}

// ReadECSParams parses the ecs-params.yml file and puts it into an ECSParams struct.
func ReadECSParams(filename string) (*ECSParams, error) {
	if filename == "" {
		defaultFilename := "ecs-params.yml"
		if _, err := os.Stat(defaultFilename); err == nil {
			filename = defaultFilename
		} else {
			return nil, nil
		}
	}

	// NOTE: Readfile reads all data into memory and closes file. Could
	// eventually refactor this to read different sections separately.
	ecsParamsData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading file '%v'", filename)
	}
	ecsParamsData = []byte(os.ExpandEnv(string(ecsParamsData)))
	ecsParams := &ECSParams{}

	if err = yaml.Unmarshal([]byte(ecsParamsData), &ecsParams); err != nil {
		return nil, errors.Wrapf(err, "Error unmarshalling yaml data from ECS params file: %v", filename)
	}

	return ecsParams, nil
}

// ConvertToECSNetworkConfiguration extracts out the NetworkConfiguration from
// the ECSParams into a format that is compatible with ECSClient calls.
func ConvertToECSNetworkConfiguration(ecsParams *ECSParams) (*ecs.NetworkConfiguration, error) {
	if ecsParams == nil {
		return nil, nil
	}

	networkMode := ecsParams.TaskDefinition.NetworkMode

	if networkMode != "awsvpc" {
		return nil, nil
	}

	awsvpcConfig := ecsParams.RunParams.NetworkConfiguration.AwsVpcConfiguration

	subnets := awsvpcConfig.Subnets

	if len(subnets) < 1 {
		return nil, errors.New("at least one subnet is required in the network configuration")
	}

	securityGroups := awsvpcConfig.SecurityGroups
	assignPublicIp := string(awsvpcConfig.AssignPublicIp)

	ecsSubnets := make([]*string, len(subnets))
	for i, subnet := range subnets {
		ecsSubnets[i] = aws.String(subnet)
	}

	ecsSecurityGroups := make([]*string, len(securityGroups))
	for i, sg := range securityGroups {
		ecsSecurityGroups[i] = aws.String(sg)
	}

	ecsAwsVpcConfig := &ecs.AwsVpcConfiguration{
		Subnets:        ecsSubnets,
		SecurityGroups: ecsSecurityGroups,
	}

	// For tasks launched with network config in EC2 mode, assign_pubic_ip field is not accepted
	if assignPublicIp != "" {
		ecsAwsVpcConfig.AssignPublicIp = aws.String(assignPublicIp)
	}

	ecsNetworkConfig := &ecs.NetworkConfiguration{
		AwsvpcConfiguration: ecsAwsVpcConfig,
	}

	return ecsNetworkConfig, nil
}
