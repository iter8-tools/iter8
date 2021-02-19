---
template: overrides/main.html
---

# spec.strategy.testingPattern

<!-- `spec.strategy.testingPattern` determines the logic used to evaluate the app versions and determine the `winner` of the experiment. iter8 supports two testing patterns, namely, `canary` and `conformance`.

- Canary: Two app versions, namely, `baseline` and `candidate`, are evaluated. Candidate is declared the winner if it satisfies experiment objectives. If candidate fails to satisfy objectives but baseline does, then baseline is declared the winner.

- Conformance: A single app version is evaluated; it is declared the winner if it satisfies experiment objectives.

The sample experiment above uses the canary testing pattern.

??? note "Links to in-depth description and code samples"
    1. In-depth description of testing patterns is [here](aspects/testing.md).
    2. Code samples... -->
