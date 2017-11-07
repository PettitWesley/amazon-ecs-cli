// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//      http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package ami

import "fmt"

// ECSAmiIds interface is used to get the ami id for a specified region.
type ECSAmiIds interface {
	Get(string) (string, error)
}

// staticAmiIds impmenets the ECSAmiIds interface to get the AMI id for
// a region using a hardcoded map of values.
type staticAmiIds struct {
	regionToId map[string]string
}

func NewStaticAmiIds() ECSAmiIds {
	regionToId := make(map[string]string)
	// amzn-ami-2017.03.g-amazon-ecs-optimized AMIs
	regionToId["us-east-1"] = "ami-20ff515a"
	regionToId["us-east-2"] = "ami-b0527dd5"
	regionToId["us-west-1"] = "ami-b388b4d3"
	regionToId["us-west-2"] = "ami-3702ca4f"
	regionToId["ca-central-1"] = "ami-fc5fe798"
	regionToId["cn-north-1"] = "ami-ea5a8987"
	regionToId["eu-central-1"] = "ami-ebfb7e84"
	regionToId["eu-west-1"] = "ami-d65dfbaf"
	regionToId["eu-west-2"] = "ami-ee7d618a"
	regionToId["ap-northeast-1"] = "ami-95903df3"
	regionToId["ap-northeast-2"] = "ami-70d0741e"
	regionToId["ap-southeast-1"] = "ami-c8c98bab"
	regionToId["ap-southeast-2"] = "ami-e3b75981"

	return &staticAmiIds{regionToId: regionToId}
}

func (c *staticAmiIds) Get(region string) (string, error) {
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}
