CREATE KEYSPACE IF NOT EXISTS rabbit WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};

CREATE TABLE IF NOT EXISTS rabbit.users
(
    url  text,
    name text,
    html text,
    PRIMARY KEY (name)
);