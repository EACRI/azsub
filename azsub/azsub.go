package azsub

// package azsub handles the creation of a node pool
// the creation of a batch job, and the submission
// of the batch task to the azure batch account

import (
	"errors"
	"fmt"
	"os"

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
	Task                *AzSubTask
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

func (a *Azsub) WithTask(task *AzSubTask) *Azsub {
	// TODO: add task definition
	a.Task = task
	return a
}

// WithStorage attempts to use default credentials
// to initialize the storage container and storage blob clients
// will exit with error if this authenticaiton fails
func (a *Azsub) WithStorage(accountName, containerPrefix string) *Azsub {

	cred, err := container.NewSharedKeyCredential(accountName, containerPrefix)
	if err != nil {
		log.Errorln("error: storage account key credential invalid")
		os.Exit(1)
	}

	containerURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerPrefix)
	cc, err := container.NewClientWithSharedKeyCredential(containerURL, cred, nil)
	if err != nil {
		log.Println(errors.New("error: storage account authentication failed"))
		os.Exit(1)
	}

	// set container client
	a.Clients.Container = cc

	return a
}

func (a *Azsub) Run() error {

	// runlocal
	if a.Local {
		return RunLocal(a)
	}

	// send the job to run on batch
	err := RunBatch(a)
	if err != nil {
		return err
	}

	return err
}
