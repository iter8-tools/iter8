# Dockerizing Iter8

Iter8 experiments can be performed a remote applications. In particular, Iter8 can be run in a local cluster while the application resides in a remote K8s cluster. In these situations, it is desirable to run Iter8 within a local cluster like Minikube or Kind which itself is containerized within a Docker image.

The Dockerfile in this folder creates such an image.