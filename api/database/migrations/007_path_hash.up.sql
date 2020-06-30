-- Update database to hash indexed paths

CREATE PROCEDURE MigratePathHashIfNeeded()
BEGIN

-- Add path hash for photo table if it doesn't exist
IF NOT EXISTS( SELECT *
            FROM INFORMATION_SCHEMA.COLUMNS
           WHERE table_name = 'photo'
			 AND table_schema = DATABASE()
             AND column_name = 'path_hash')  THEN

    -- Remove unique index from photo.path
    ALTER TABLE photo DROP INDEX path;

    -- Add path_hash and set it to the md5 hash based of the path attribute
    ALTER TABLE photo ADD path_hash varchar(32) AFTER path;
    UPDATE photo p SET path_hash = md5(p.path);
    ALTER TABLE photo MODIFY path_hash varchar(32) NOT NULL UNIQUE;

END IF;

-- Add path hash for album table if it doesn't exist
IF NOT EXISTS( SELECT *
            FROM INFORMATION_SCHEMA.COLUMNS
           WHERE table_name = 'album'
			 AND table_schema = DATABASE()
             AND column_name = 'path_hash')  THEN

    -- Remove unique index from album.path
    ALTER TABLE album DROP INDEX path;

    -- Add path_hash and set it to the md5 hash based of the path attribute
    ALTER TABLE album ADD path_hash varchar(32) AFTER path;
    UPDATE album a SET path_hash = md5(a.path);
    ALTER TABLE album MODIFY path_hash varchar(32) NOT NULL UNIQUE;

END IF;

END; -- MigratePathHashIfNeeded procedure end

CALL MigratePathHashIfNeeded();
DROP PROCEDURE MigratePathHashIfNeeded;