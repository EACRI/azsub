package azsub

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/batch/armbatch"
	"github.com/Azure/go-autorest/autorest/to"
	log "github.com/sirupsen/logrus"
)

const (
	DEFAULT_BATCH_POOL_IMAGE_PUBLISHER      = "microsoft-azure-batch"
	DEFAULT_BATCH_POOL_IMAGE_OFFER          = "ubuntu-server-container"
	DEFAULT_BATCH_POOL_IMAGE_SKU            = "20-04-lts"
	DEFAULT_BATCH_POOL_IMAGE_VERSION        = "latest"
	DEFAULT_BATCH_POOL_NODE_SIZE            = "Standard_D2_v3"
	DEFAULT_BATCH_POOL_NODE_AGENT_SKU_ID    = "batch.node.ubuntu 20.04"
	DEFAULT_BATCH_POOL_NODE_CONTAINER_IMAGE = "ubuntu"
	DEFAULT_BATCH_POOL_NODE_COUNT           = 1
	DEFAULT_TASK_SLOTS_PER_NODE             = 1
	DEFAULT_TARGET_DEDICATED_NODES          = 1
)

// BatchPoolNodeConfig structures variables
// for the batch pool configuration
type BatchPoolConfig struct {
	Publisher            string
	Sku                  string
	Offer                string
	Version              string
	SkuID                string
	Container            string
	NodeSize             string
	NodeCount            int32
	TaskSlotsPerNode     int32
	TargetDedicatedNodes int32
	StartTask            *BatchPoolStartTask
	Clients              *Clients
}

// BatchPoolStartTask configures the startTask
// on the batch pool nodes
type BatchPoolStartTask struct {
	SourceUrl  string
	FilePath   string
	StdOutFile string
}

type ArmBatchClients struct {
	Pool *armbatch.PoolClient
}

func NewBatchPoolConfig() *BatchPoolConfig {
	return &BatchPoolConfig{
		Publisher:            DEFAULT_BATCH_POOL_IMAGE_PUBLISHER,
		Sku:                  DEFAULT_BATCH_POOL_IMAGE_SKU,
		Offer:                DEFAULT_BATCH_POOL_IMAGE_OFFER,
		Version:              DEFAULT_BATCH_POOL_IMAGE_VERSION,
		NodeSize:             DEFAULT_BATCH_POOL_NODE_SIZE,
		SkuID:                DEFAULT_BATCH_POOL_NODE_AGENT_SKU_ID,
		Container:            DEFAULT_BATCH_POOL_NODE_CONTAINER_IMAGE,
		NodeCount:            DEFAULT_BATCH_POOL_NODE_COUNT,
		TaskSlotsPerNode:     DEFAULT_TASK_SLOTS_PER_NODE,
		TargetDedicatedNodes: DEFAULT_TARGET_DEDICATED_NODES,
		StartTask:            nil,
	}

}

// createBatchPool creates an Azure Batch compute node pool
func (b *BatchPoolConfig) createBatchPool(ctx context.Context, resourceGroupID, subscriptionID, accountName, accountLocation, poolID string) error {

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	poolClient, _ := armbatch.NewPoolClient(subscriptionID, cred, nil)

	elevation := armbatch.ElevationLevelAdmin
	scope := armbatch.AutoUserScopeTask

	// define node pool creation settings
	pool := &armbatch.Pool{
		ID: &poolID,
		Properties: &armbatch.PoolProperties{
			DisplayName: to.StringPtr("azsub"),
			StartTask: &armbatch.StartTask{
				CommandLine: to.StringPtr("bash -f start.sh"),
				ResourceFiles: []*armbatch.ResourceFile{
					{
						FileMode:   to.StringPtr("777"),
						BlobPrefix: to.StringPtr("/path/to/blob/prefix"),
						FilePath:   to.StringPtr("start.sh"),
					},
				},
				UserIdentity: &armbatch.UserIdentity{
					AutoUser: &armbatch.AutoUserSpecification{
						ElevationLevel: &elevation,
						Scope:          &scope,
					},
				},
				WaitForSuccess: to.BoolPtr(true),
			},
			VMSize:           to.StringPtr(b.NodeSize),
			TaskSlotsPerNode: to.Int32Ptr(b.TaskSlotsPerNode),
			DeploymentConfiguration: &armbatch.DeploymentConfiguration{
				VirtualMachineConfiguration: &armbatch.VirtualMachineConfiguration{
					ImageReference: &armbatch.ImageReference{
						Offer:     to.StringPtr(b.Offer),
						Publisher: to.StringPtr(b.Publisher),
						SKU:       to.StringPtr(b.Sku),
						Version:   to.StringPtr(b.Version),
					},
					NodeAgentSKUID:         to.StringPtr(b.SkuID),
					ContainerConfiguration: &armbatch.ContainerConfiguration{},
				},
			},
		},
	}

	res, err := poolClient.Create(ctx, resourceGroupID, accountName, poolID, *pool, nil)
	if err != nil {
		log.Errorln(err)
	}

	fmt.Println(res)
	return nil
}

func (b *BatchPoolConfig) createBatchJob() error { return nil }

// run batch send the task definition to batch execution
func RunBatch(a *Azsub) error {

	// b := NewBatchPoolConfig()
	// err := b.createBatchPool()
	// err := b.createBatchJob()
	// err := b.createBatchTask()

	return nil
}
