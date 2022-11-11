# Azsub

Simple command line interface for submitting one off tasks to azure batch.

## Motivation

Azure batch allows you to submit compute tasks to ephemeral compute resources managed by Azure. The service is flexible and elastic, but the user experience requires multi-step setup, monitoring and takedown. Azsub aims to simplify the user experience by providing a command-line interface that submits tasks to an existing batch account, while abstracting intermediate steps such as node pool, job, and task creation. The resources are destroyed when the task Succeeds, Fails or Exits unexpectedly.