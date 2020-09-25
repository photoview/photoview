
ALTER TABLE site_info
  ADD COLUMN IF NOT EXISTS periodic_scan_interval int(8) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS concurrent_workers int(8) NOT NULL DEFAULT 3;
