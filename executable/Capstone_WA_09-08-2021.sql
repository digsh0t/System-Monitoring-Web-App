CREATE DATABASE CP_Server_Administrator_WA;
USE CP_Server_Administrator_WA;

CREATE TABLE wa_users (
  wa_users_id INT PRIMARY KEY AUTO_INCREMENT,
  wa_users_username VARCHAR(60),
  wa_users_password VARCHAR(130),
  wa_users_role VARCHAR(60)
);

INSERT INTO wa_users (wa_users_username, wa_users_password, wa_users_role) VALUES ("trilx123","9a835b7eece9ea09bfc80b63d15b94aee929eac524544813da1962bc35081fbaea7698c84b73b7b3d7c65ead23d7abbf0d8e25e183e50f6a1f1e96f97d712afd", "admin");

SELECT * FROM wa_users;

CREATE TABLE  ssh_keys (
    sk_key_id INT PRIMARY KEY AUTO_INCREMENT,
    sk_key_name varchar(60),
    sk_private_key text,
    creator_id INT,
    FOREIGN KEY (creator_id) references wa_users(wa_users_id)
);

CREATE TABLE ssh_connections (
    sc_connection_id INT PRIMARY KEY AUTO_INCREMENT,
    sc_username VARCHAR(60),
    sc_host varchar(60),
    sc_port INT,
    creator_id INT,
    ssh_key_id INT,
    FOREIGN KEY (creator_id) references wa_users(wa_users_id),
    FOREIGN KEY (ssh_key_id) references ssh_keys(sk_key_id)
);

ALTER TABLE ssh_connections ADD sc_password varchar(50);

SELECT * FROM wa_users;
