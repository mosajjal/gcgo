package ai

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

func TestCommandTreeIncludesOperations(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))
	var operationsFound, deployFound, undeployFound, datasetsFound, pipelineJobsFound, batchPredictionJobsFound bool

	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "operations":
			operationsFound = true
		case "endpoints":
			for _, nested := range sub.Commands() {
				switch nested.Name() {
				case "deploy-model":
					deployFound = true
				case "undeploy-model":
					undeployFound = true
				}
			}
		case "datasets":
			datasetsFound = true
		case "pipeline-jobs":
			pipelineJobsFound = true
		case "batch-prediction-jobs":
			batchPredictionJobsFound = true
		}
	}
	if !operationsFound || !deployFound || !undeployFound || !datasetsFound || !pipelineJobsFound || !batchPredictionJobsFound {
		t.Fatalf("operations=%v deploy=%v undeploy=%v datasets=%v pipelineJobs=%v batchPredictionJobs=%v", operationsFound, deployFound, undeployFound, datasetsFound, pipelineJobsFound, batchPredictionJobsFound)
	}
}
