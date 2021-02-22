---
template: overrides/main.html
---

# spec.strategy.deploymentPattern

<!-- `spec.strategy.deploymentPattern` is a string enum that determines if and how traffic is shifted during an experiment with the `Canary` testing pattern. iter8 supports two deployment patterns, namely, `progressive` and `fixed-split`.

- Progressive: Progressively shift traffic towards the winner during each iteration of the experiment.

- Fixed-split: The traffic split set at the start of the experiment is left unchanged during iterations. -->
