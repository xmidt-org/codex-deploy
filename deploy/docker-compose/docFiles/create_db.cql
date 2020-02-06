CREATE KEYSPACE IF NOT EXISTS devices;
CREATE TABLE devices.events (device_id  varchar,
    record_type INT,
    birthdate BIGINT,
    deathdate BIGINT,
    data BLOB,
    nonce BLOB,
    alg VARCHAR,
    kid VARCHAR,
    row_id TIMEUUID,
    PRIMARY KEY (device_id, birthdate, record_type))
    WITH CLUSTERING ORDER BY (birthdate DESC, record_type ASC)
    AND default_time_to_live = 300
    AND transactions = {'enabled': 'false'};
CREATE INDEX search_by_record_type ON devices.events
    (device_id, record_type, birthdate) 
    WITH CLUSTERING ORDER BY (record_type ASC, birthdate DESC)
    AND default_time_to_live = 300
    AND transactions = {'enabled': 'false', 'consistency_level':'user_enforced'};
CREATE INDEX search_by_row_id ON devices.events
    (device_id, row_id) 
    WITH CLUSTERING ORDER BY (row_id DESC)
    AND default_time_to_live = 300
    AND transactions = {'enabled': 'false', 'consistency_level':'user_enforced'};
CREATE TABLE devices.blacklist (device_id varchar PRIMARY KEY, reason varchar);
