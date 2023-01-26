/*
    UserDefaultMapping.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "UserDefaultMapping.h"

std::string UserDefaultMapping::getId() {    
    return this->id;
}

void UserDefaultMapping::setId(std::string &value) {
    this->id = value;
}

std::string UserDefaultMapping::getUserId() {    
    return this->userId;
}

void UserDefaultMapping::setUserId(std::string &value) {
    this->userId = value;
}

std::string UserDefaultMapping::getUsername() {    
    return this->username;
}

void UserDefaultMapping::setUsername(std::string &value) {
    this->username = value;
}

std::string UserDefaultMapping::getSecret() {    
    return this->secret;
}

void UserDefaultMapping::setSecret(std::string &value) {
    this->secret = value;
}

int UserDefaultMapping::getCreated() {    
    return this->created;
}

void UserDefaultMapping::setCreated(int &value) {
    this->created = value;
}

int UserDefaultMapping::getModified() {    
    return this->modified;
}

void UserDefaultMapping::setModified(int &value) {
    this->modified = value;
}

bool UserDefaultMapping::isDeleted() {    
    return this->deleted;
}

void UserDefaultMapping::setDeleted(bool &value) {
    this->deleted = value;
}

