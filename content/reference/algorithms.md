---
menuTitle: Algorithms
title: Algorithms
weight: 63
summary: Coming soon!
---

Iter8 approaches continuous experimentation (canary release, A/B and A/B/n rollouts) as an online iterative decision making problem. At the start of every iteration of an experiment, iter8 computes assessments for each version based on all currently available observations, uses these assessments to rank the versions, and decides how best to control the split of traffic between them using its traffic control strategy.

An in-depth description of the ML foundations of iter8's algorithms is available in [this USENIX HotCloud'20 paper](https://www.usenix.org/conference/hotcloud20/presentation/toslali). Below, we present the essentials of iter8's version assessments and traffic control strategies.

## Version Assessments
Iter8 uses an online Bayesian learning approach for continually assessing how each version is performing with respect to each metric and relative to each other, based on the all data observed until the current point in time. Iter8 surfaces a variety of useful insights based on this Bayesian assessment during the course of the experiment and after its termination. They include probability of a version being the winner, the probability of a version improving over the baseline with respect to a given metric, the probability of a version being the best version with respect to a given metric, the range of values that a version is likely to take on for a given metric (credible interval).

## Traffic Control Strategies

Iter8 provides three distinct traffic control strategies, namely, `progressive`, `top_2`, and `uniform`, with `progressive` being the default and recommended strategy for common scenarios. All three strategies are built on a statistically robust machine learning (ML) foundation that employs Bayesian estimation for assessing versions and a multi-armed bandit algorithm for deciding traffic splits. The three strategies are described in the following table.

Strategy | Description | When to use
---------|-------------|-------------
*progressive* | Progressively shift all traffic to the winner. | This is the default strategy. Use this when you doing canary releases or A/B or A/B/n rollouts and your goal is safely and reliably shift all traffic to the winning version. 
*top_2* | Converge towards a 50-50 traffic split between the best two versions | This strategy helps find the winning version in fewer iterations compared to `progressive`. Use this if your goal is not progressive traffic shifting but quick winner identification.
*uniform* | Converge towards a uniform traffic split across all versions. | Use this strategy when your goal is to maximize what you learn about each version in the experiment. In particular, this strategy is useful if you want sharper credible intervals for each metric for each version of the experiment.