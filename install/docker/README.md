# Iter8 in Docker

Iter8 experiments can be performed for remote applications. In particular, Iter8 can be run in a local cluster while the application resides in a remote K8s cluster. In these situations, it is desirable to run Iter8 within a local cluster like Minikube or Kind that is containerized within a Docker image.

The Dockerfile in this folder creates such an image.

## Building Iter8-in-Docker locally

1. Export ITER8 and cd
```shell
export ITER8=<path-to-root-of-your-local-iter8-repo>
cd $ITER8/install/docker
```

2. Build ind image
```shell
docker build -t ind:latest .
```

## Running the Iter8-in-Docker container

> If you want to run using the local image, replace iter8/ind:latest with ind:latest below.

> If you want to run a specific version, replace iter8/ind:latest with iter8/ind:<tag> below. For example, `iter8/ind:0.7.21`.

1. Start Iter8 container
```shell
docker run --name ind --privileged -d iter8/ind:latest
```
To pin the version of Iter8, replace `latest` above with tag (for example, with `0.7.21`)

2. Setup Iter8 within container
```shell
docker exec ind ./iter8.sh
```

3. Run conformance test for your application
```shell
docker exec ind helm install \
--set URL=https://example.com \
--set LimitMeanLatency='"200.0"' \
--set LimitErrorRate='"0.01"' \
--set Limit95thPercentileLatency='"500.0"' \
codeengine /iter8/helm/conformance
```

4. Describe results of the conformance test
```shell
docker exec ind \
watch -n 10.0 \
"kubectl get experiment codeengine-experiment -o yaml | iter8ctl describe -f -"
```

5. Remove Iter8-in-Docker container and image
```shell
docker rm -f -v ind
docker rmi -f iter8/ind:latest
```