
50 VUs:

DIRECT REDIS: 
redis-benchmark -h 127.0.0.1 -p 6379 -n 1000 -c 50
Summary:
  throughput summary: 41666.67 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        0.799     0.184     0.703     1.423     2.583     3.103

Aegis:
redis-benchmark -h 127.0.0.1 -p 6380 -n 1000 -c 50

Summary:
  throughput summary: 16666.67 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        2.743     0.576     2.415     5.255     7.719     8.751


--------------------------------------------------------------------------



Only Get/set
redis-benchmark -t get,set -n 10000 -p 6379 -c 50

Redis: 
Summary:
  throughput summary: 43478.26 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        0.606     0.144     0.559     0.911     1.455     2.079


AEGIS: only get set

Summary:
  throughput summary: 21413.28 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        2.020     0.352     1.823     3.535     5.671    12.303


------------------------------------------------------------------------


100 VUs:

REDIS get/set/del

siuumanth@Victuss:~$ redis-benchmark -t get,set,del -n 10000 -p 6380 -c 100

Summary:
  throughput summary: 42735.04 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        1.228     0.472     1.119     1.831     2.639     3.487
        

AEGIS get/set/del all enabled:

Summary:
  throughput summary: 26737.97 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        2.932     0.496     2.511     5.511     9.407    14.111


---------------------------------------------------------------------

AEGIS specific:
with 4 hot key workers

100 VUs:
hot keys and tags disabled
Summary:
  throughput summary: 27700.83 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        2.829     0.552     2.519     5.191     6.263     8.575

tags disabled, hot keys enabled
Summary:
  throughput summary: 23752.97 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        3.392     0.664     2.807     7.175    16.255    21.503

conclusion: mutex contention very high 
options:
1. remove mutex
2. sharded maps 

5 workers:
Summary:
  throughput summary: 23640.66 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        3.271     0.632     2.799     6.743     9.239    15.727


with 20 hot key workers 
test 1:
Summary:
  throughput summary: 28169.02 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        2.786     0.488     2.463     5.167     8.631    22.655

test 2:
Summary:
  throughput summary: 26954.18 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        3.931     0.560     3.119     8.839    17.327    29.327


------------------------


NEW TESTS:

5 workers, channel size 1000, mutex
Summary:
  throughput summary: 22727.27 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        3.507     0.592     2.959     7.215    12.063    22.559

5 workers, channel size 10000, mutex

Summary:
  throughput summary: 22471.91 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        3.428     0.624     2.919     7.143    13.975    21.711

Summary:
  throughput summary: 28735.63 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        2.751     0.760     2.463     4.727    10.143    15.591

20 workers, channel size 10000, mutex

Summary:
  throughput summary: 24752.47 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        3.267     0.528     2.775     6.223    12.903    23.311

5 workers, 10000 size, and mutex in hot keys disabled
1.
Summary:
  throughput summary: 27322.40 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        2.924     0.664     2.543     5.295    11.071    14.399

2.
Summary:
  throughput summary: 27322.40 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        2.924     0.664     2.543     5.295    11.071    14.399


20 workers, 10000 size, mutex disabled

Summary:
  throughput summary: 22573.36 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        3.599     0.840     3.103     6.927    11.607    17.231


All features false:
Aegis
Summary:
  throughput summary: 27397.26 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        2.863     0.632     2.591     4.775     9.719    14.279





Pure redis for benchmark:
Summary:
  throughput summary: 42016.80 requests per second
  latency summary (msec):
          avg       min       p50       p95       p99       max
        1.226     0.328     1.127     1.711     2.487     4.759



