package azsub

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/batch/batch"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/gofrs/uuid"
)

const (
	DEFAULT_BATCH_POOL_IMAGE_PUBLISHER      = "microsoft-azure-batch"
	DEFAULT_BATCH_POOL_IMAGE_OFFER          = "ubuntu-server-container"
	DEFAULT_BATCH_POOL_IMAGE_SKU            = "20-04-lts"
	DEFAULT_BATCH_POOL_IMAGE_VERSION        = "latest"
	DEFAULT_BATCH_POOL_IMAGE_SIZE           = "Standard_D2_v3"
	DEFAULT_BATCH_POOL_NODE_AGENT_SKU_ID    = "batch.node.ubuntu 20.04"
	DEFAULT_BATCH_POOL_NODE_CONTAINER_IMAGE = "ubuntu"
	DEFAULT_BATCH_POOL_NODE_COUNT           = 1
)

// BatchPoolNodeConfig structures variables
// for the batch pool configuration
type BatchPoolNodeConfig struct {
	Publisher string
	Sku       string
	Offer     string
	Version   string
	Size      string
	SkuID     string
	Container string
	NodeCount int
	StartTask *BatchPoolStartTask
}

// BatchPoolStartTask configures the startTask
// on the batch pool nodes
type BatchPoolStartTask struct {
	SourceUrl  string
	FilePath   string
	StdOutFile string
}

func NewBatchPoolNodeConfig() *BatchPoolNodeConfig {
	return &BatchPoolNodeConfig{
		Publisher: DEFAULT_BATCH_POOL_IMAGE_PUBLISHER,
		Sku:       DEFAULT_BATCH_POOL_IMAGE_SKU,
		Offer:     DEFAULT_BATCH_POOL_IMAGE_OFFER,
		Version:   DEFAULT_BATCH_POOL_IMAGE_VERSION,
		Size:      DEFAULT_BATCH_POOL_IMAGE_SIZE,
		SkuID:     DEFAULT_BATCH_POOL_NODE_AGENT_SKU_ID,
		Container: DEFAULT_BATCH_POOL_NODE_CONTAINER_IMAGE,
		NodeCount: DEFAULT_BATCH_POOL_NODE_COUNT,
		StartTask: nil,
	}

}

func getAccountClient() batchARM.AccountClient {
	accountClient := batchARM.NewAccountClient(config.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer()
	accountClient.Authorizer = auth
	_ = accountClient.AddToUserAgent(config.UserAgent())
	return accountClient
}

func getPoolClient(accountName, accountLocation string) batch.PoolClient {
	poolClient := batch.NewPoolClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	auth, _ := iam.GetBatchAuthorizer()
	poolClient.Authorizer = auth
	_ = poolClient.AddToUserAgent(config.UserAgent())
	poolClient.RequestInspector = fixContentTypeInspector()
	return poolClient
}

func getJobClient(accountName, accountLocation string) batch.JobClient {
	jobClient := batch.NewJobClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	auth, _ := iam.GetBatchAuthorizer()
	jobClient.Authorizer = auth
	_ = jobClient.AddToUserAgent(config.UserAgent())
	jobClient.RequestInspector = fixContentTypeInspector()
	return jobClient
}

func getTaskClient(accountName, accountLocation string) batch.TaskClient {
	taskClient := batch.NewTaskClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	auth, _ := iam.GetBatchAuthorizer()
	taskClient.Authorizer = auth
	_ = taskClient.AddToUserAgent(config.UserAgent())
	taskClient.RequestInspector = fixContentTypeInspector()
	return taskClient
}

func getFileClient(accountName, accountLocation string) batch.FileClient {
	fileClient := batch.NewFileClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	auth, _ := iam.GetBatchAuthorizer()
	fileClient.Authorizer = auth
	_ = fileClient.AddToUserAgent(config.UserAgent())
	fileClient.RequestInspector = fixContentTypeInspector()
	return fileClient
}

// create batch node pool, batch jobs to node pool, batch task to batch job
// CreateBatchPool creates an Azure Batch compute pool
func CreateBatchPool(ctx context.Context, accountName, accountLocation, poolID string) error {
	poolClient := getPoolClient(accountName, accountLocation)
	toCreate := batch.PoolAddParameter{
		ID: &poolID,
		VirtualMachineConfiguration: &batch.VirtualMachineConfiguration{
			ImageReference: &batch.ImageReference{
				Publisher: to.StringPtr("Canonical"),
				Sku:       to.StringPtr("16.04-LTS"),
				Offer:     to.StringPtr("UbuntuServer"),
				Version:   to.StringPtr("latest"),
			},
			NodeAgentSKUID: to.StringPtr("batch.node.ubuntu 16.04"),
		},
		MaxTasksPerNode:      to.Int32Ptr(1),
		TargetDedicatedNodes: to.Int32Ptr(1),
		// Create a startup task to run a script on each pool machine
		StartTask: &batch.StartTask{
			ResourceFiles: &[]batch.ResourceFile{
				{
					BlobSource: to.StringPtr("https://raw.githubusercontent.com/lawrencegripper/azure-sdk-for-go-samples/1441a1dc4a6f7e47c4f6d8b537cf77ce4f7c452c/batch/examplestartup.sh"),
					FilePath:   to.StringPtr("echohello.sh"),
					FileMode:   to.StringPtr("777"),
				},
			},
			CommandLine:    to.StringPtr("bash -f echohello.sh"),
			WaitForSuccess: to.BoolPtr(true),
			UserIdentity: &batch.UserIdentity{
				AutoUser: &batch.AutoUserSpecification{
					ElevationLevel: batch.Admin,
					Scope:          batch.Task,
				},
			},
		},
		VMSize: to.StringPtr("standard_a1"),
	}

	_, err := poolClient.Add(ctx, toCreate, nil, nil, nil, nil)

	if err != nil {
		return fmt.Errorf("cannot create pool: %v", err)
	}

	return nil
}

// CreateBatchJob create an azure batch job
func createBatchJob(ctx context.Context, accountName, accountLocation, poolID, jobID string) error {
	jobClient := getJobClient(accountName, accountLocation)
	jobToCreate := batch.JobAddParameter{
		ID: to.StringPtr(jobID),
		PoolInfo: &batch.PoolInformation{
			PoolID: to.StringPtr(poolID),
		},
	}
	_, err := jobClient.Add(ctx, jobToCreate, nil, nil, nil, nil)

	if err != nil {
		return err
	}

	return nil
}

// createBatchTask create an azure batch job
func createBatchTask(ctx context.Context, accountName, accountLocation, jobID string) (string, error) {
	taskID, _ := uuid.NewV4()
	taskIDString := taskID.String()
	taskClient := getTaskClient(accountName, accountLocation)
	taskToAdd := batch.TaskAddParameter{
		ID:          &taskIDString,
		CommandLine: to.StringPtr("/bin/bash -c 'set -e; set -o pipefail; echo Hello world from the Batch Hello world sample!; wait'"),
		UserIdentity: &batch.UserIdentity{
			AutoUser: &batch.AutoUserSpecification{
				ElevationLevel: batch.Admin,
				Scope:          batch.Task,
			},
		},
	}
	_, err := taskClient.Add(ctx, jobID, taskToAdd, nil, nil, nil, nil)

	if err != nil {
		return "", err
	}

	return taskIDString, nil
}

// waitForTaskResult polls the task and retreives it's stdout once it has completed
func waitForTaskResult(ctx context.Context, accountName, accountLocation, jobID, taskID string) (stdout string, err error) {
	taskClient := getTaskClient(accountName, accountLocation)
	res, err := taskClient.Get(ctx, jobID, taskID, "", "", nil, nil, nil, nil, "", "", nil, nil)
	if err != nil {
		return "", err
	}
	waitCtx, cancel := context.WithTimeout(ctx, time.Minute*4)
	defer cancel()

	if res.State != batch.TaskStateCompleted {
		for {
			_, ok := waitCtx.Deadline()
			if !ok {
				return stdout, errors.New("timedout waiting for task to execute")
			}
			time.Sleep(time.Second * 15)
			res, err = taskClient.Get(ctx, jobID, taskID, "", "", nil, nil, nil, nil, "", "", nil, nil)
			if err != nil {
				return "", err
			}
			if res.State == batch.TaskStateCompleted {
				waitCtx.Done()
				break
			}
		}
	}

	fileClient := getFileClient(accountName, accountLocation)

	reader, err := fileClient.GetFromTask(ctx, jobID, taskID, stdoutFile, nil, nil, nil, nil, "", nil, nil)

	if err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadAll(*reader.Value)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func getBatchBaseURL(accountName, accountLocation string) string {
	return fmt.Sprintf("https://%s.%s.batch.azure.com", accountName, accountLocation)
}
