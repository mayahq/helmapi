package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func getChartInfoFromRuntimeId(runtimeId string) (map[string]interface{}, error) {
	app := "helm"
	instanceName := "rt-" + runtimeId
	args := []string{"get", "values", instanceName, "-o", "json"}

	if len(runtimeId) == 0 {
		return nil, fmt.Errorf("you cannot provide an empty runtime ID")
	}

	cmd := exec.Command(app, args...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var output map[string]interface{}
	err = json.NewDecoder(&outb).Decode(&output)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func GetInstallRequestFromRuntimeId(runtimeId string) (InstallRequest, error) {
	instanceDetails, err := getChartInfoFromRuntimeId(runtimeId)
	if err != nil {
		return InstallRequest{}, err
	}

	var ir InstallRequest
	ir.ChartName = "mayanr"
	ir.ReleaseName = "rt-" + runtimeId
	ir.PrivateChartsRepo = instanceDetails["privateChartsRepo"].(string)
	ir.Values = instanceDetails
	ir.Flags = []string{}

	return ir, nil
}
func GetDeleteRequestFromRuntimeId(runtimeId string) (DeleteRequest, error) {
	// Doing this to make sure that the runtime exists
	_, err := getChartInfoFromRuntimeId(runtimeId)
	if err != nil {
		return DeleteRequest{}, err
	}

	var dr DeleteRequest
	dr.ReleaseName = "rt-" + runtimeId

	return dr, nil
}
