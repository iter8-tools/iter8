# Use Iter8 in a full CI/CD pipeline on Openshift

This tutorial describes how one can make use of Iter8 to enhance an existing CI/CD pipeline to perform progressive rollout. There are 2 parts to this tutorial. First part walks through how one sets up a fully functional CI/CD pipeline using Openshift Pipeline and Openshift GitOps running in an Openshift cluster. This allows developer's code changes in Git repo to trigger new image build and later deploy into a cluster, i.e., a GitOps experience. Second part assumes the user already has a fully functional CI/CD pipeline and now wants to leverage Iter8 to perform additional testing before rolling out a new version.

## Part 1: Set up a CI/CD pipeline on Openshift

### Openshift cluster

You will need an Openshift cluster. We are using an Openshift cluster running on [IBM Cloud](https://www.ibm.com/cloud) for this tutorial. You can also use [CodeReady Container](https://developers.redhat.com/products/codeready-containers/overview) or [Openshift Playground](https://developers.redhat.com/courses/openshift/playground-openshift) to follow along.

Once you have an Openshift cluster, the following components need to be installed
- Openshift Pipeline
- Openshift GitOps

### Install Openshift Pipeline

From IBM Cloud Console > Openshift > Cluster > [my cluster] > Openshift web console > Operators > OperatorHub > Red Hat OpenShift Pipelines Operator > Install. Once it's installed, you can check the Openshift Pipeline Operator and the Pipeline components are running in the cluster by running the following commands:

```shell
oc get pods -n openshift-operators
NAME                                            READY   STATUS    RESTARTS   AGE
openshift-pipelines-operator-7fdd8fff9f-sv5dj   1/1     Running   0          4m3s

oc get pods -n openshift-pipelines
NAME                                          READY   STATUS    RESTARTS   AGE
tekton-pipelines-controller-7c4b9bf4b-gmdpk   1/1     Running   0          75s
tekton-pipelines-webhook-6f666f55f7-s9wcv     1/1     Running   0          75s
tekton-triggers-controller-54f8c88b4b-lmrdb   1/1     Running   0          39s
tekton-triggers-webhook-c64fd9b47-pvtk8       1/1     Running   0          39s
```

You can check Openshift Pipeline Web Console by:

### Install Openshift GitOps

From IBM Cloud Console > Openshift > Cluster > [my cluster] > Openshift web console > Operators > OperatorHub > Red Hat OpenShift GitOps > Install. Once it's installed, you can check the Openshift GitOps Operator and GitOps components are running in the cluster by running the following commands:

```shell
oc get pods -n openshift-operators
NAME                                            READY   STATUS    RESTARTS   AGE
gitops-operator-6bccf5bbdf-xcfkx                1/1     Running   0          17s
openshift-pipelines-operator-7fdd8fff9f-sv5dj   1/1     Running   0          4m3s

oc -n openshift-gitops get pods
NAME                                                    READY   STATUS    RESTARTS   AGE
argocd-cluster-application-controller-7fbb6d4f6-sv5q6   0/1     Running   0          2m
argocd-cluster-redis-74cc6c9f46-bslhp                   1/1     Running   0          2m
argocd-cluster-repo-server-65f74dddbb-pjmw5             1/1     Running   0          119s
argocd-cluster-server-84b95d5cc-29wlf                   0/1     Running   0          119s
kam-7974577cdc-hz5kb                                    1/1     Running   0          2m3s
```

You can access Openshift GitOps Web Console by clicking Application Stages > ArgoCD from the Openshift Web Console directly, or you can do a `port-forward` locally, i.e.,

```shell
oc port-forward -n openshift-gitops svc/argocd-cluster-server 8080:443
```

and then open a web browser to http://localhost:8080. In either case, you will need an `admin` password to login to ArgoCD Web Console, and it can be obtained with the following command:

```shell
oc -n openshift-gitops  get secret argocd-cluster-cluster -o jsonpath="{.data.admin\.password}" | base64 -d
```

### Setup Github token

Components in the CI/CD pipeline will need to modify your Github repo via creating PRs, and you need to provide authentication to it by creating a Github Personal Access Token. Go to github.com > upper right corner > Settings > Developer settings > Personal access token > Generate new token

Copy the token and make a Kubernetes secret from it so it can be used at runtime by the CI/CD pipeline.

```shell
oc create secret generic github-token --from-literal=token=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Initialize and deploy your app

### Repo setup

Since this is a GitOps scenario, you will need to use your own Github repo. You can fork this repository: [https://github.com/iter8-tools/iter8](https://github.com/iter8-tools/iter8) so it is available at https://github.com/[YOUR_ORG]/iter8. Once this is done, you can clone it locally:

```shell
git clone https://github.com/[YOUR_ORG]/iter8.git
cd iter8
export ITER8=$(pwd)
$ITER8/samples/gitops/platformsetup.sh
```

replacing [YOUR_ORG] with your Github organization or username. Now, do the same replacement operation to update some references in the repo so they will point at your forked repo.

```shell
find $ITER8/samples/gitops -name "*" -type f | xargs sed -i '' "s/MY_ORG/YOUR_ORG/"
git commit -a -m "update reference links"
git push origin head
```

### Deploy app 

```shell
oc apply -f $ITER8/samples/cicd/rbac.yaml
oc apply -f $ITER8/samples/cicd/argocd-app.yaml
```

Now Argo CD Web Console should show that a new app called `gitops` is created. Make sure it is showing both Healthy and Synced - this might take a few minutes.

## Setup Github webhook

When a developer makes a change in the repo that would require a new image to be built, we want Github to send a webhook call so it triggers pipeline tools to start building the image. We first need to setup Openshift Pipeline to be ready to receive the webhook calls, and then we will go to Github to configure the sending side of the webhook call.

### Setup Openshift Pipeline

Run the following commands:

```shell
oc apply -f $ITER8/samples/cicd/tekton/eventlistener.yaml
oc apply -f $ITER8/samples/cicd/tekton/pipeline.yaml
oc apply -f $ITER8/samples/cicd/tekton/pipelineresources.yaml
oc apply -f $ITER8/samples/cicd/tekton/triggerbinding.yaml
oc apply -f $ITER8/samples/cicd/tekton/triggertemplate.yaml
oc apply -f $ITER8/samples/cicd/tekton/tasks/
```

You will also need to expose the EventListener service via Openshift Route so Github can send webhook calls to the cluster. This can be done via

```
oc expose service el-iter8-eventlistener
```

### Setup Github webhook

Go to github.com/[YOUR_ORG]/iter8 > Settings > Webhooks > Add webhook. To fill in `Payload URL`, run the following command:

```shell
oc get route el-iter8-eventlistener --template='http://{{.spec.host}}'
```

Set `Content-type` to `application/json`. Set `Which event` to `Let me select individual event`, and then select `Pull request` and unselect `Push`.

## Start a progressive rollout

Now that our application is deployed in the cluster, CI/CD pipeline configured to watch for commits from Github, it is time to see what happens when a developer merges a PR.


