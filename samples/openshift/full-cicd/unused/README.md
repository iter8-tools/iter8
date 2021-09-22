istio/: These files are needed when Istio is deployed in an Openshift cluster using its built-in Prometheus Operator, so that Prometheus instance will scrape Istio sidecars properly.
metrics.yaml: These are Openshift metrics collected from Openshift Route
route.yaml: This is an Openshift Route used to shift traffic
fortio.yaml: This is a workload generator 
