
ALTER TABLE media
  DROP COLUMN IF EXISTS side_car_path,
  DROP COLUMN IF EXISTS side_car_hash;
