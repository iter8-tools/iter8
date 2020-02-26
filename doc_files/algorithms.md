# iter8's algorithms

This documentation briefly describes the algorithms supported by iter8 to make decisions during canary releases or A/B testing. These algorithms are part of iter8's analytics service (_iter8-analytics_) and exposed via REST API. Iter8's Kubernetes controller (_iter8-controller_) calls the appropriate REST API based on the `.spec.trafficControl.strategy` set in a custom `Experiment` resource. Iter8's `Experiment` CRD is documented [here](iter8_crd.md).

Iter8's algorithms are statistically robust. Below, we list the algorithms currently available to users.  This list will grow as we introduce other sophisticated algorithms based on Bayesian strategies for decision making.

## 1. Progressive check-and-increment algorithm (`check_and_increment`)

#### Input parameters

```yaml
interval: # (time; e.g., 30s)
maxIterations: # (integer; e.g., 1000)
trafficStepSize: # (percentage; e.g., 5)
maxTrafficPercentage: # 80 (percentage; e.g., 90)
onSuccess: # (string enum; possible values are: "candidate", "baseline", "both")
```

This algorithm is suitable for the gradual rollout of a candidate ("canary") version. The goal of this strategy is to gradually shift traffic from a baseline (stable) version to a candidate version, as long as the candidate version continues to pass the success criteria defined by the user. 

When the `experiment` begins, the traffic split is as follows: `trafficStepSize`% to the candidate version, and `100 - trafficStepSize`% to the baseline version. At the end of each iteration (whose duration is determined by the `interval` parameter), iter8 checks if there are enough data points to decide whether the candidate version satisfies the success criteria (i.e, whether enough requests were sent to make a statistically robust assessment). If there is enough data to make a decision, and the candidate version satisfies all criteria, iter8 increases the traffic to the candidate version by `trafficStepSize`. Else, if there is insufficient data, or if the candidate version fails to satisfy one or more success criteria, then the traffic split does not change. Furthermore, if a failing criterion has been declared by the user as critical, iter8 aborts the experiment and makes sure all traffic goes to the baseline version. This is a rollback situation.

A successful experiment will last for a duration of length  `interval * maxIterations`. In case of success, the user can specify whether iter8 should: (1) send all traffic to the candidate; (2) roll back to the baseline despite success; or (3) split traffic across both versions. If the traffic is to be split across both versions, then the final split will be as follows: `maxTrafficPercentage`% to the candidate and `1 - maxTrafficPercentage` to the baseline.

## 2. Decaying epsilon-greedy algorithm (`epsilon_greedy`)

```yaml
interval: # (time; e.g., 30s)
maxIterations: # (integer; e.g., 1000)
maxTrafficPercentage: # 80 (percentage; e.g., 90)
onSuccess: # (string enum; possible values are: "candidate", "baseline", "both")
```

This algorithm can be applied to canary releases as well as A/B or A/B/n testing. The goal of this strategy is to explore two or more competing versions, aiming to maximize a reward (which is typically associated with business-oriented metrics) while making sure that the defined success criteria (typically associated with performance-oriented and/or correctness-oriented metrics) are satisfied.

Unlike the check-and-increment strategy described above, this algorithm automatically decides what the proper traffic split should be at the end of each iteration and does not require the user to supply a value for the traffic increment per iteration. It converges relatively quickly to the "optimal" version as more iterations occur over time.

In A/B or A/B/n testing, the "optimality" of a version relates to maximizing the reward during the course of an experiment while satisfying the success criteria. In the context of canary releases, an implicit reward metric is used to indicate whether or not the success criteria are satisfied at each iteration.

## 3. Posterior Bayesian Routing (`pbr`)

One of our novel algorithms. Coming soon to iter8...

## 4. Optimistic Bayesian Routing (`obr`)

One of our novel algorithms. Coming soon to iter8...
