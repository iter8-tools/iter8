---
menuTitle: Request Header Matching
title: Request Header Matching
weight: 10
summary: Learn how to experiment on a subset of requests defined by a request headers
---

You may wish to experiment on a subset of all user traffic -- that traffic that matches certain conditions.
Remaining traffic should continue to be handled by the existing baseline version.
At the end of the experiment, all of the traffic should be reconfigured to the winner, if one is selected, or the baseline otherwise.
This task describes how you can experiment on the subset of requests defined by a match on the user request headers.

## Define Match Rule

A rule defining the matching traffic on which experiments should be executed is defined in the `trafficControl` section of an `Experiment`. Here is an example:

```yaml
match:
  http:
    - headers:
        user:
          exact: john-doe
```

In this example, test traffic will be restricted to those requests that contain a request header `User: john-doe`.
Such traffic will be split between the baseline and candidate versions according to the recommendation of the analytics iter8 carries out.
Any non-matching traffic will be sent only to the baseline version.
