CREATE TABLE wa_users (wa_users_id INT PRIMARY KEY AUTO_INCREMENT,wa_users_username VARCHAR(60),wa_users_password VARCHAR(130),wa_users_name VARCHAR(60),wa_users_role VARCHAR(60));

CREATE TABLE  ssh_keys (sk_key_id INT PRIMARY KEY AUTO_INCREMENT,sk_key_name varchar(60),sk_private_key text,creator_id INT,FOREIGN KEY (creator_id) references wa_users(wa_users_id));

CREATE TABLE ssh_connections (sc_connection_id INT PRIMARY KEY AUTO_INCREMENT,sc_username VARCHAR(60),sc_password varchar(50),sc_host varchar(60),sc_hostname varchar(60),sc_port INT,creator_id INT,ssh_key_id INT,FOREIGN KEY (creator_id) references wa_users(wa_users_id),FOREIGN KEY (ssh_key_id) references ssh_keys(sk_key_id));

CREATE TABLE package_installed (pkg_id INT PRIMARY KEY AUTO_INCREMENT,pkg_name VARCHAR(60),pkg_date DATETIME,pkg_host_id INT,FOREIGN KEY (pkg_host_id) references ssh_connections(sc_connection_id) ON DELETE CASCADE);

CREATE TABLE event_web (ev_web_id INT PRIMARY KEY AUTO_INCREMENT,ev_web_type VARCHAR(60),ev_web_description VARCHAR(300),ev_web_timestamp DATETIME,ev_web_creator_id INT,FOREIGN KEY (ev_web_creator_id) references wa_users(wa_users_id));

CREATE TABLE  invent_group (invent_group_id INT PRIMARY KEY AUTO_INCREMENT,invent_group_name varchar(60));

ALTER TABLE ssh_connections ADD group_id INT;
ALTER TABLE ssh_connections ADD FOREIGN KEY (group_id) references invent_group(invent_group_id);


CREATE TABLE snmp_credential (snmp_id INT PRIMARY KEY AUTO_INCREMENT,snmp_auth_username varchar(60),snmp_auth_password varchar(60),snmp_priv_password varchar(60),snmp_connection_id INT,FOREIGN KEY (snmp_connection_id) references ssh_connections(sc_connection_id) ON DELETE CASCADE);

CREATE TABLE ssh_connections_information (sc_info_id INT PRIMARY KEY AUTO_INCREMENT,sc_info_osname varchar(60),sc_info_osversion varchar(60),sc_info_installdate DATETIME,sc_info_serial varchar(60),sc_info_connection_id int,FOREIGN KEY (sc_info_connection_id) references ssh_connections(sc_connection_id) ON DELETE CASCADE);


