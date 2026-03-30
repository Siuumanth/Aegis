**on resume features** — what you have is already strong. but things that would make it pop:

- **metrics endpoint** — expose hit rate, miss rate, hot key count, singleflight collapse ratio via a simple HTTP `/metrics` endpoint. Prometheus format. interviewers love observability
- **README with benchmarks** — `redis-benchmark` results, before/after latency, singleflight collapse numbers under load. concrete numbers on a resume > feature lists
- **config hot reload** — watch `config.yaml` for changes, reload policies without restart. shows you understand production ops
- **CLI flag** — `--config path/to/config.yaml` instead of hardcoded path. tiny but looks professional