-- Check if the user exists, then create it
CREATE USER kwild WITH PASSWORD 'kwild' SUPERUSER REPLICATION;
CREATE DATABASE kwild OWNER kwild;