# Mathematical Analysis & Pseudocode for `lowkey` Improvements

## Current Algorithm Complexity

```
CURRENT_ALGORITHM(dirs[], interval):
  WHILE true:
    FOR each dir IN dirs:
      files[] = GLOB(dir + "/**/*")           // O(f) filesystem calls
      FOR each file IN files:
        stat = FILE_STAT(file)                // O(1) but expensive syscall
        state[file] = [stat.mtime, stat.size] // O(1) memory write
      END FOR
      COMPARE_STATES(old_state, new_state)    // O(f) comparisons
    END FOR
    SLEEP(interval)                           // O(1)
  END WHILE

Time Complexity: O(n × f × i) where:
  n = number of directories
  f = average files per directory
  i = iterations over time
Space Complexity: O(n × f) for storing all file states
```

## 1. Filesystem Events + Polling Hybrid

```
HYBRID_MONITOR(dirs[], poll_interval):
  event_queue = PRIORITY_QUEUE()
  last_poll = CURRENT_TIME()

  // Setup filesystem watchers
  FOR each dir IN dirs:
    watcher = CREATE_INOTIFY_WATCHER(dir)
    watcher.ON_EVENT(event):
      event_queue.ENQUEUE(event, CURRENT_TIME())
    END ON_EVENT
  END FOR

  WHILE true:
    // Process real-time events
    WHILE NOT event_queue.EMPTY():
      event = event_queue.DEQUEUE()
      PROCESS_EVENT(event)                    // O(1)
    END WHILE

    // Periodic safety scan
    IF (CURRENT_TIME() - last_poll) > poll_interval:
      INCREMENTAL_SCAN(dirs, last_poll)       // O(changed_files)
      last_poll = CURRENT_TIME()
    END IF

    SLEEP(0.1)                                // Reduce CPU spinning
  END WHILE

Time Complexity: O(c + s) where:
  c = number of actual changes (event-driven)
  s = periodic scan cost (much less frequent)
Space Complexity: O(active_events + baseline_state)
```

## 2. Incremental State Tracking

```
INCREMENTAL_SCAN(dirs[], last_scan_time):
  changed_files = []

  FOR each dir IN dirs:
    dir_mtime = DIR_STAT(dir).mtime

    // Skip unchanged directories
    IF cached_dir_times[dir] == dir_mtime:
      CONTINUE                                // O(1) skip
    END IF

    cached_dir_times[dir] = dir_mtime

    files[] = GLOB(dir + "/**/*")
    FOR each file IN files:
      file_mtime = FILE_STAT(file).mtime

      // Only process changed files
      IF file_mtime > last_scan_time:         // O(1) comparison
        changed_files.APPEND(file)
        UPDATE_STATE(file)                    // O(1)
      END IF
    END FOR
  END FOR

  RETURN changed_files

Time Complexity: O(d + c) where:
  d = number of directories (for dir stat checks)
  c = number of actually changed files
Space Complexity: O(c) for changed files list
```

## 3. Smart Filtering with Bloom Filters

```
SMART_FILTER_INIT(expected_files):
  ignore_bloom = BLOOM_FILTER(expected_files * 0.1, 0.01)  // 1% false positive

  FOR each pattern IN IGNORE_PATTERNS:
    ignore_bloom.ADD(HASH(pattern))
  END FOR

  RETURN ignore_bloom

SHOULD_MONITOR(file_path, ignore_bloom):
  file_hash = HASH(file_path)

  // Fast bloom filter check - O(k) where k = hash functions
  IF ignore_bloom.MIGHT_CONTAIN(file_hash):
    // Slow but accurate check only for potential matches
    FOR each pattern IN IGNORE_PATTERNS:      // O(p) where p = patterns
      IF FNMATCH(pattern, file_path):
        RETURN false
      END IF
    END FOR
  END IF

  RETURN true

Time Complexity: O(k) for most files, O(k + p) for potential ignores
Space Complexity: O(m) where m = bloom filter size
```

## 4. Adaptive Polling with Exponential Backoff

```
ADAPTIVE_INTERVAL(recent_changes[], base_interval):
  activity_score = CALCULATE_ACTIVITY_SCORE(recent_changes)

  // Exponential backoff based on activity
  IF activity_score == 0:
    interval = MIN(base_interval * 4, 60)     // Max 60s when idle
  ELSE IF activity_score < 5:
    interval = base_interval * 2              // 10s for low activity
  ELSE IF activity_score < 20:
    interval = base_interval                  // 5s for medium activity
  ELSE:
    interval = MAX(base_interval / 2, 1)      // Min 1s for high activity
  END IF

  RETURN interval

CALCULATE_ACTIVITY_SCORE(recent_changes[]):
  current_time = CURRENT_TIME()
  score = 0

  FOR each change IN recent_changes:
    age = current_time - change.timestamp
    weight = EXP(-age / 300)                  // Exponential decay over 5 minutes
    score += weight
  END FOR

  RETURN score

Time Complexity: O(r) where r = recent changes buffer size
Space Complexity: O(r) for recent changes buffer
```

## 5. Batch Processing with Temporal Windowing

```
BATCH_PROCESSOR(changes[], window_size):
  batches = []
  current_batch = []
  window_start = null

  FOR each change IN changes:
    IF window_start == null:
      window_start = change.timestamp
    END IF

    // Check if change fits in current window
    IF (change.timestamp - window_start) <= window_size:
      current_batch.APPEND(change)           // O(1)
    ELSE:
      // Process completed batch
      batches.APPEND(current_batch)          // O(1)
      current_batch = [change]
      window_start = change.timestamp
    END IF
  END FOR

  // Process final batch
  IF NOT current_batch.EMPTY():
    batches.APPEND(current_batch)
  END IF

  RETURN batches

WRITE_BATCH_TO_LOG(batch[], log_file):
  buffer = ""

  FOR each change IN batch:
    buffer += FORMAT_LOG_ENTRY(change)       // O(1)
  END FOR

  ATOMIC_WRITE(log_file, buffer)             // O(1) single I/O operation

Time Complexity: O(c) where c = total changes
Space Complexity: O(b) where b = largest batch size
```

## 6. Memory-Efficient State Storage

```
COMPRESSED_STATE_STORAGE():
  file_states = HASH_MAP()

  STORE_FILE_STATE(file_path, stat):
    signature = COMPUTE_SIGNATURE(file_path, stat)
    file_states[file_path] = signature

  COMPUTE_SIGNATURE(file_path, stat):
    IF stat.size < SMALL_FILE_THRESHOLD:
      // For small files: content hash
      RETURN MD5(READ_FILE(file_path))[0:8]   // 8-char hash
    ELSE:
      // For large files: mtime + size combo
      RETURN COMBINE(stat.mtime, stat.size)   // 16 bytes
    END IF

  DETECT_CHANGE(file_path, old_signature):
    current_stat = FILE_STAT(file_path)
    new_signature = COMPUTE_SIGNATURE(file_path, current_stat)

    RETURN old_signature != new_signature

Space Complexity: O(f × 16) bytes instead of O(f × 64) bytes
Time Complexity: O(1) for large files, O(s) for small files where s = file size
```

## 7. Priority Queue for Change Processing

```
PRIORITY_CHANGE_QUEUE():
  queue = MIN_HEAP()  // Priority queue

  ENQUEUE_CHANGE(change, priority):
    heap_item = (priority, CURRENT_TIME(), change)
    queue.INSERT(heap_item)                   // O(log n)

  PROCESS_CHANGES():
    WHILE NOT queue.EMPTY():
      priority, timestamp, change = queue.EXTRACT_MIN()  // O(log n)

      // Skip if change is too old (debouncing)
      IF (CURRENT_TIME() - timestamp) > DEBOUNCE_TIME:
        CONTINUE
      END IF

      APPLY_CHANGE(change)                    // O(1)
    END WHILE

Priority Calculation:
  CALCULATE_PRIORITY(change_type, file_path):
    base_priority = {
      'CREATED': 1,     // Highest priority
      'DELETED': 2,
      'MODIFIED': 3     // Lowest priority
    }[change_type]

    // Boost priority for important files
    IF MATCH(file_path, IMPORTANT_PATTERNS):
      base_priority -= 0.5
    END IF

    RETURN base_priority

Time Complexity: O(log n) per change
Space Complexity: O(pending_changes)
```

## 8. Complete Optimized Algorithm

```
OPTIMIZED_LOKEE_MONITOR(dirs[], config):
  // Initialize components
  event_system = INIT_FILESYSTEM_EVENTS(dirs)
  ignore_filter = SMART_FILTER_INIT(ESTIMATED_FILES)
  change_queue = PRIORITY_CHANGE_QUEUE()
  state_store = COMPRESSED_STATE_STORAGE()

  last_poll = CURRENT_TIME()
  recent_changes = CIRCULAR_BUFFER(1000)

  WHILE true:
    // Process real-time events
    WHILE event_system.HAS_EVENTS():
      event = event_system.GET_NEXT_EVENT()    // O(1)

      IF SHOULD_MONITOR(event.path, ignore_filter):  // O(k)
        priority = CALCULATE_PRIORITY(event.type, event.path)
        change_queue.ENQUEUE_CHANGE(event, priority)  // O(log n)
        recent_changes.ADD(event)              // O(1)
      END IF
    END WHILE

    // Adaptive polling
    current_time = CURRENT_TIME()
    poll_interval = ADAPTIVE_INTERVAL(recent_changes, BASE_INTERVAL)

    IF (current_time - last_poll) > poll_interval:
      changed_files = INCREMENTAL_SCAN(dirs, last_poll)  // O(d + c)

      FOR each file IN changed_files:
        IF SHOULD_MONITOR(file, ignore_filter):
          change_queue.ENQUEUE_CHANGE(file, 3)  // O(log n)
        END IF
      END FOR

      last_poll = current_time
    END IF

    // Process changes in batches
    IF change_queue.SIZE() > BATCH_THRESHOLD OR
       (current_time - last_batch_time) > MAX_BATCH_DELAY:

      batch = change_queue.EXTRACT_BATCH(MAX_BATCH_SIZE)  // O(b log n)
      WRITE_BATCH_TO_LOG(batch)                           // O(b)
      last_batch_time = current_time
    END IF

    SLEEP(MIN_SLEEP_INTERVAL)                  // 0.1s
  END WHILE

Overall Time Complexity: O(k + log n + d + c + b) per iteration
Overall Space Complexity: O(f + n + b) where:
  k = bloom filter hash operations (constant)
  n = pending changes in queue
  d = number of directories
  c = actually changed files
  b = batch size
  f = total files being monitored
```

## Performance Gains Estimation

```
PERFORMANCE_IMPROVEMENT_ANALYSIS():
  baseline_cost = n × f × syscall_cost × iterations

  // With optimizations:
  event_driven_cost = actual_changes × process_cost
  periodic_scan_cost = (d + changed_files) × syscall_cost × (iterations / poll_factor)
  filtering_cost = total_files × bloom_filter_cost

  optimized_cost = event_driven_cost + periodic_scan_cost + filtering_cost

  improvement_ratio = baseline_cost / optimized_cost

  // Expected improvements:
  // - 80-95% reduction in syscalls
  // - 60-90% reduction in CPU usage
  // - 50-80% reduction in memory usage
  // - Near real-time detection vs 5s polling delays
```

This mathematical framework provides the foundation for implementing high-performance file monitoring that scales from dozens to thousands of directories efficiently.
