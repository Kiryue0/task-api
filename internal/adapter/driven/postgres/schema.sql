-- Uygulama açılışında otomatik çalıştırılır (idempotent);
-- istenirse elle de uygulanabilir: psql -f schema.sql
--
-- Kullanıcının oluşturduğu tüm tablolar bu şemada durur. Böylece uygulamanın
-- veya Postgres'in kendi tablolarıyla asla karışmazlar.
CREATE SCHEMA IF NOT EXISTS user_tables;
