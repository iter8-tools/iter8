# Iter8 in Docker

Iter8 experiments can be performed for remote applications. In particular, Iter8 can be run in a local cluster while the application resides in a remote K8s cluster. In these situations, it is desirable to run Iter8 within a local cluster like Minikube or Kind that is containerized within a Docker image.

The Dockerfile in this folder creates such an image.

## Building and running Iter8 in Docker locally

1. Export ITER8 and cd
```shell
export ITER8=<path-to-root-of-your-local-iter8-repo>
cd $ITER8/install/docker
```

2. Build image
```shell
docker build -t iter8:ind .
```

3. Run image in detached mode
```shell
docker run --name ind --privileged -d iter8:ind
```

4. Run Iter8 in Docker
```shell
docker exec ind /iter8/iter8.sh
```

5. Cleanup
Cleanup remnants after finishing Step 4 (or before Step 3 if needed).
```shell
docker kill ind
docker rm ind
```