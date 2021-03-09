??? info "Understanding what happened"
    1. You configured three Knative services corresponding to two versions of your app in `services.yaml`. (probably using kubectl?)

    2. You used `customdomain.com` as the HTTP host in this tutorial.
        - **Note:** In your production cluster, use domain(s) that you own in the setup of the virtual service.

    3. You set up an Istio virtual service which mapped the Knative services to this custom domain. The virtual service specified the following routing rules: all HTTP requests to `customdomain.com` with their Host header or :authority pseudo-header **not** set to `wakanda` would be routed to the `baseline`; those with `wakanda` Host header or :authority pseudo-header may be routed to `baseline` and `candidate`.
    
    4. The percentage of `wakandan` requests sent to `candidate` is 0% at the beginning of the experiment.

    5. You generated traffic for `customdomain.com` using a `curl`-job with two `curl`-containers to simulate user requests. You injected Istio sidecar injected into it to simulate traffic generation from within the cluster. The sidecar was needed in order to correctly route traffic. One of the `curl`-containers sets the `country` header field to `wakanda`, and the other to `gondor`.
        - **Note:** You used Istio version 1.8.2 to inject the sidecar. This version of Istio corresponds to the one installed in [Step 3 of the quick start tutorial](http://localhost:8000/getting-started/quick-start/with-knative/#3-install-knative-and-iter8). If you have a different version of Istio installed in your cluster, change the Istio version during sidecar injection appropriately.
    
    6. You created an Iter8 `Canary` experiment with `Progressive` deployment pattern to evaluate the `candidate`. In each iteration, Iter8 observed the mean latency, 95th percentile tail-latency, and error-rate metrics collected by Prometheus, and verified that the `candidate` version satisfied all the `objectives` specified in the experiment. It progressively increased the proportion of traffic with `country: wakanda` header that is routed to the `candidate`.
