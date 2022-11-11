package azsub

// package azsub handles the creation of a node pool
// the creation of a batch job, and the submission
// of the batch task to the azure batch account

import (
	"fmt"

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
	return a
}

// WithStorageAccount attempts to use default credentials
// to initialize the container client and the blob client
func (a *Azsub) WithStorageAccont(accountName string) *Azsub {

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Errorln(err)
	}

	client, err := azblob.NewClient(serviceURL, cred, nil)
	if err != nil {
		log.Errorln(err)
	}

	a.Clients.Blob = client

	return a
}

func (a *Azsub) WithContainer(containerName string) *Azsub {

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Errorln(err)
	}

	containerUrl := fmt.Sprintf("https://%s", containerName)
	cc, err := container.NewClient(containerUrl, cred, &container.ClientOptions{})
	if err != nil {
		log.Errorln(err)
	}

	a.Clients.Container = cc

	return a
}

func (a *Azsub) Run() error {

	// authenticate
	err := authenticate()

	// name jobs

	// runlocal
	if a.Local {
		runlocal()
	}

	// create batch pool
	pool, err := createBatchPool()
	if err != nil {
		log.Errorln(err)
		return err
	}

	// create batch job
	job, err := pool.CreateBatchJob()
	if err != nil {
		log.Errorln(err)
		return err
	}

	// create batch task
	err := job.CreateTask()
	if err != nil {
		log.Errorln(err)
		return err
	}
}
