
ALTER TABLE site_info
  DROP COLUMN IF EXISTS periodic_scan_interval,
  DrOP COLUMN IF EXISTS concurrent_workers;
