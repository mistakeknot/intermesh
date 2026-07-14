# Intermesh optimization campaigns

| Campaign | Retained result | Holdout result | Verdict |
|---|---|---|---|
| [intermesh-abstention-v1](intermesh-abstention-v1/learnings.md) | Development no-match recall 0.20 → 0.7333; positive metrics unchanged | No-match recall 0.05 → 0.40; top-3/top-5 unchanged at 0.60/0.70 | Keep two stopword mutations; revert the route-level evidence gate |
| [intermesh-routing-v2](intermesh-routing-v2/learnings.md) | c03-derived top-3 recall 0.00 → 1.00; development metrics unchanged | No-match recall 0.40, top-3 0.60, top-5 0.70, and MRR 0.4558 unchanged | Keep identifier-boundary exact-name matching |
| [intermesh-routing-v3](intermesh-routing-v3/learnings.md) | c04-derived MRR 0.3333 → 1.00; development metrics unchanged | No-match recall 0.40, top-3 0.60, top-5 0.70, and MRR 0.4558 unchanged | Keep direct-negation-aware lexical matching |
