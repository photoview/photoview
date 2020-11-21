
ALTER TABLE site_info
  DROP COLUMN IF EXISTS periodic_scan_interval,
  DROP COLUMN IF EXISTS concurrent_workers;
