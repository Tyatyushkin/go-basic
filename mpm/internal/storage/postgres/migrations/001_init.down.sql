-- Удаление триггеров
DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;
DROP TRIGGER IF EXISTS update_photos_updated_at ON photos;
DROP TRIGGER IF EXISTS update_albums_updated_at ON albums;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Удаление функции
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаление индексов
DROP INDEX IF EXISTS idx_photo_metadata_photo_id;
DROP INDEX IF EXISTS idx_comments_user_id;
DROP INDEX IF EXISTS idx_comments_photo_id;
DROP INDEX IF EXISTS idx_photos_hash;
DROP INDEX IF EXISTS idx_photos_created_at;
DROP INDEX IF EXISTS idx_photos_user_id;
DROP INDEX IF EXISTS idx_photos_album_id;
DROP INDEX IF EXISTS idx_albums_created_at;
DROP INDEX IF EXISTS idx_albums_user_id;

-- Удаление таблиц в правильном порядке (учитывая зависимости)
DROP TABLE IF EXISTS photo_metadata;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS photo_tags;
DROP TABLE IF EXISTS album_tags;
DROP TABLE IF EXISTS photos;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS albums;
DROP TABLE IF EXISTS users;