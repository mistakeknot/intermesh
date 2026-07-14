# Ideas backlog

## Tried

- [x] Treat `by` as a function-word stopword — retained; c08-derived
  no-match recall and precision improved from 0 to 1.
- [x] Add a conservative IDF signal for query terms found inside compound
  skill names — retained at 0.9x; c07-derived top-three recall improved from
  0.5 to 1 while development MRR improved.

## Rejected

- A 1.5x compound-name weight restored observed recall but reduced development
  MRR from 0.8417 to 0.8122.
- A 0.8x compound-name weight improved development MRR but left the expected
  c07 review skill at rank four.
