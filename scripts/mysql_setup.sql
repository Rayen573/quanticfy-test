-- Create database and application user (run inside MySQL CLI)
CREATE DATABASE IF NOT EXISTS quanticfy_test;
CREATE USER IF NOT EXISTS 'quanticfy'@'localhost' IDENTIFIED BY 'YOUR_PASSWORD';
GRANT ALL PRIVILEGES ON quanticfy_test.* TO 'quanticfy'@'localhost';
FLUSH PRIVILEGES;