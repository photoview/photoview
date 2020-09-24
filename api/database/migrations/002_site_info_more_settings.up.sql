
ALTER TABLE site_info
  ADD COLUMN IF NOT EXISTS periodic_scan_interval int(8) NOT NULL,
  ADD COLUMN IF NOT EXISTS concurrent_workers int(8) NOT NULL;
