## Current Version Features

## 0.1.2 - AI help
1. **Flagging (Generalistic analysis)**: Flagging potential issues or inneficiencies.

- A report based on the pprof of a profile.
  - **Info To Include**:
    1.  bench/tag/text/BenchmarkName_profile.txt (all)
    2.  bench/tag/profile_functions_BenchmarkName_profile.png (all)

1. **Structure**:
   bench/Tag/AI/generalistic/BenchmarkName
   generalistic_analysis.txt
   generalistic_analysis.txt

## Generalistic Analysis Guide

1. What is the resource consumption breakdown between the library, benchmark function, and profiling framework? (Include percentages and absolute values)
2. What's the activity detected when it comes to the GC? (e.g. time, memory)
3. What are the key functions from the library consuming the most resources, and what is their ratio of flat time (direct execution) to cumulative time (including called functions)?
4. What does the runtime activity tell about your system?
5. What is the distribution of time spent in system calls vs user code, and how does this pattern change across different parts of the benchmark execution?
6. Are there any unexpected high-frequency function calls that might indicate inefficiencies?
7. Are there any synchronization primitives (locks, channels) that show significant contention?
8. Are there any patterns in the call stack that suggest potential optimization opportunities?
9. Write a general analysis flagging areas that could be improved.
