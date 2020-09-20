
ALTER TABLE site_info
  ADD COLUMN IF NOT EXISTS periodic_scan_interval int(8) NOT NULL DEFAULT 0;
