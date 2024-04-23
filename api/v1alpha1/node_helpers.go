package v1alpha1

// validate the Node and NodeProperties passed are valid
// e.g validate vmNames being unique, properties correctly defined, etc.
// one day move to common/util

import (
	"fmt"
)

var (
	typeOptions     = map[string]bool{"database": true, "haproxy": true}
	roleOptions     = map[string]bool{"primary": true, "secondary": true}
	failoverOptions = map[string]bool{"Automatic": true, "Manual": true}
)

func ValidateNodes(nodes []*Node, isHighAvailability bool) error {
	if !isHighAvailability {
		return nil
	}
	databaseNodeCount := 0
	vmNames := make(map[string]bool) // for validating that vmnames are unique
	for _, node := range nodes {
		if node.Properties.NodeType == "database" {
			databaseNodeCount++
		}
		if err := ValidateNodeProperties(node.Properties); err != nil {
			return err
		}

		if _, ok := vmNames[node.VmName]; ok {
			return fmt.Errorf("vmName %s is already specified", node.VmName)
		}
		vmNames[node.VmName] = true
	}
	if databaseNodeCount < 3 {
		return fmt.Errorf("high Availability requires at least 3 nodes database nodes")
	}
	return nil
}

func ValidateNodeProperties(np NodeProperties) error {
	if !typeOptions[np.NodeType] {
		return fmt.Errorf("invalid NodeType in Node Properties: %s", np.NodeType)
	}

	if !roleOptions[np.Role] {
		return fmt.Errorf("invalid Role in Node Properties: %s", np.Role)
	}

	if !failoverOptions[np.FailoverMode] {
		return fmt.Errorf("invalid FailoverMode in Node Properties: %s", np.FailoverMode)
	}

	return nil
}
