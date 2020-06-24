-- Migrate all tables in database to use utf8 for better language support
ALTER TABLE access_token CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
ALTER TABLE album CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
ALTER TABLE photo CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
ALTER TABLE photo_exif CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
ALTER TABLE photo_url CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
ALTER TABLE share_token CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
ALTER TABLE site_info CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
ALTER TABLE user CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
