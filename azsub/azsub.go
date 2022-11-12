package azsub

// package azsub handles the creation of a node pool
// the creation of a batch job, and the submission
// of the batch task to the azure batch account

import (
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/services/batch/2019-08-01.10.0/batch"
	log "github.com/sirupsen/logrus"
)

// define interface for task builders to implement
// type AzTaskBuilder interface {
// 	WithLocal(string) *AzTaskBuilder,
// 	WithImage() *AzTaskBuilder,
// 	WithContainer() *AzTaskBuilder,
// }

type Clients struct {
	Blob      *azblob.Client
	Container *container.Client
	Batch     *batch.AccountClient
}

// Azsub struct builds an azsub submission
type Azsub struct {
	Jobid               string
	ContainerImage      string
	BlobContainerPrefix string
	Clients             Clients
	Local               bool
	Task                AzSubTask
	StartTaskUrl        string
}

// TaskType Enum for AzSubTask definition
type TaskType int

const (
	CommandTypeTask TaskType = iota
	ScriptTypeTask
)

type AzSubTask struct {
	Type TaskType
	Task string
}

func (t *AzSubTask) IsCommand() bool {
	return t.Type == CommandTypeTask
}

func (t *AzSubTask) IsScript() bool {
	return t.Type == ScriptTypeTask
}

func NewAzsub() *Azsub {
	return &Azsub{Jobid: "newjobid"}
}

// configure container image to use
// with batch jobs
func (a *Azsub) WithImage(image string) *Azsub {
	a.ContainerImage = image
	return a
}

func (a *Azsub) WithLocal(local bool) *Azsub {
	a.Local = local
	return a
}

func (a *Azsub) WithTask(AzSubTask) *Azsub {
	// TODO: add task definition
	return a
}

// WithStorage attempts to use default credentials
// to initialize the storage container and storage blob clients
// will exit with error if this authenticaiton fails
func (a *Azsub) WithStorage(accountName, containerPrefix string) *Azsub {

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		// TODO: attempt to use storage account key environment variable on error
		log.Errorln(err)
		os.Exit(1)
	}

	containerURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerPrefix)
	containerClient, err := container.NewClient(containerURL, cred, nil)
	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	// set container client
	a.Clients.Container = containerClient
	log.Infoln("Using container client at URL: %s", containerClient.URL())

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	blobClient, err := azblob.NewClient(serviceURL, cred, nil)
	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	// set blobClient
	a.Clients.Blob = blobClient
	log.Infoln("Using blob client at URL: %s", a.Clients.Blob.URL())

	return a
}

func (a *Azsub) Run() error {

	// runlocal
	if a.Local {
		return runlocal(a)
	}

	conf := NewBatchPoolConfig()

	err := conf.createBatchPool()

	// create batch job
	err := createBatchJob()

	// create task
	err := createTask()

}
