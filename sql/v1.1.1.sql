CREATE TABLE user (
    id INT AUTO_INCREMENT PRIMARY KEY, 
    uid INT NOT NULL,
    username VARCHAR(64) NOT NULL,
    fullname VARCHAR(128) NOT NULL,
    UNIQUE(uid, username, fullname)
);
CREATE INDEX user_uid ON user(uid);
CREATE INDEX user_username ON user(username);
CREATE INDEX user_fullname ON user(fullname);

CREATE TABLE device (
    id INT AUTO_INCREMENT PRIMARY KEY, 
    serial VARCHAR(32) NOT NULL,
    clientidentifier VARCHAR(64) NOT NULL,
    hostname VARCHAR(32) NOT NULL,
    UNIQUE(serial, clientidentifier, hostname)
);
CREATE INDEX device_serial ON device(serial);
CREATE INDEX device_clientidentifier ON device(clientidentifier);
CREATE INDEX device_hostname ON device(hostname);

CREATE TABLE address (
    id INT AUTO_INCREMENT PRIMARY KEY, 
    ip VARCHAR(15) NOT NULL,
    internetip VARCHAR(15) NOT NULL,
    UNIQUE(ip, internetip)
);
CREATE INDEX address_ip ON address(ip);
CREATE INDEX address_internetip ON address(internetip);

CREATE TABLE identity (
    id INT AUTO_INCREMENT PRIMARY KEY, 
    user_id INT NOT NULL,
    device_id INT NOT NULL,
    address_id INT NOT NULL,
    UNIQUE(user_id, device_id, address_id),
    FOREIGN KEY(user_id) REFERENCES user(id),
    FOREIGN KEY(device_id) REFERENCES device(id),
    FOREIGN KEY(address_id) REFERENCES address(id)
);

CREATE TABLE log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY, 
    identity_id INT NOT NULL,
    time DATETIME NOT NULL,
    FOREIGN KEY(identity_id) REFERENCES identity(id)
);
CREATE INDEX log_time ON log(time);
