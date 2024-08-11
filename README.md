# kubectl-label-assist
Kubernetes resource-label query helper for kubectl

## Kubernetes Resource Label Autocomplete
This feature enhances Kubernetes resource label auto-completion with a tab press. It uses the current Kubernetes context, including the API server location and authentication, to query and list labels.

It consists of two main components:

Go Code: Queries the Kubernetes API server to retrieve labels.
Bash Autocompletion Script: Uses the Go code to list labels and provide auto-completion (currently supports bash only).

_Please check the README under the completion folder for instructions on setting up auto-completion._

## How does it work?

kubectl allows for the completion of resource types and names but doesn’t support resource labels. Labels in Kubernetes aren't treated as resources in the API server—they're attributes of resources like nodes, pods, and config maps. This program queries the specific resource API to extract labels directly from the resources.

The shell completion scripts in the completion folder provide dynamic autocompletion and are configured to work with kubectl.

To optimize performance, the API calls to the Kubernetes API server use server-side printing, which reduces the amount of data sent to the client by fetching only the most relevant details. Results are cached on the file system for 5 minutes to minimize repeated API calls. You can check the .cache folder for the cached data.


## Installtion and Usage
```
cd kubectl-label-assist

go install cmd/autocomplete/kubectl-la-autocomplete.go 

source ./completion/kubectl-la.bash

kubectl get pods -n kube-system -l <tab key press>
app                         app.kubernetes.io/name      controller-revision-hash    name                        pod-template-hash
app.kubernetes.io/instance  component                   k8s-app                     pod-template-generation     tier

kubectl get pods -n kube-system -l k8s-app= <tab key press>
calico-kube-controllers  calico-node              kube-dns                 kube-proxy
```
