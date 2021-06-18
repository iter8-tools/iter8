# Iter8 in Docker

Iter8 experiments can be performed for remote applications. In particular, Iter8 can be run in a local cluster while the application resides in a remote K8s cluster. In these situations, it is desirable to run Iter8 within a local cluster like Minikube or Kind that is containerized within a Docker image.

The Dockerfile in this folder creates such an image.

## Building and running Iter8 in Docker locally

1. Export ITER8 and cd
```shell
export ITER8=<path-to-root-of-your-local-iter8-repo>
cd $ITER8/install/docker
```

2. Build ind image
```shell
docker build -t ind:latest .
```

3. Start ind container
```shell
docker run --name ind --privileged -d ind:latest
```

4. Create K8s cluster and install Iter8 -- all inside ind container
```shell
docker exec ind iter8.sh
```

To pin the version of the Dockerfile used in this image (example, `v0.6.5`), do as follows.
```shell
docker exec -e TAG=v0.6.5 ind iter8.sh
```

5. Create Iter8 experiment
```shell
docker exec ind helm install \
  --set URL=https://example.com \
  --set LimitMeanLatency='"200.0"' \
  --set LimitErrorRate='"0.01"' \
  codeengine /iter8/helm/conformance
```

6. Describe results of the experiments
```shell
docker exec ind watch \
  iter8ctl describe -f - <(kubectl get experiment conformance-experiment -o yaml)
```

7. Cleanup (do this before Step 3 if needed)
```shell
$ITER8/install/docker/cleanup.sh
```