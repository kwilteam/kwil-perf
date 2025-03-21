ALTER SYSTEM SET
 shared_buffers = '6400MB';
ALTER SYSTEM SET
 effective_cache_size = '19200MB';
ALTER SYSTEM SET
 maintenance_work_mem = '1600MB';
ALTER SYSTEM SET
 checkpoint_completion_target = '0.9';
ALTER SYSTEM SET
 wal_buffers = '16MB';
ALTER SYSTEM SET
 default_statistics_target = '100';
ALTER SYSTEM SET
 random_page_cost = '1.1';
ALTER SYSTEM SET
 effective_io_concurrency = '200';
ALTER SYSTEM SET
 work_mem = '8MB';
ALTER SYSTEM SET
 huge_pages = 'off';
ALTER SYSTEM SET
 min_wal_size = '1GB';
ALTER SYSTEM SET
 max_wal_size = '4GB';
ALTER SYSTEM SET
 max_worker_processes = '12';
ALTER SYSTEM SET
 max_parallel_workers_per_gather = '4';
ALTER SYSTEM SET
 max_parallel_workers = '12';
ALTER SYSTEM SET
 max_parallel_maintenance_workers = '4';
ALTER SYSTEM SET
 max_prepared_transactions = '2';