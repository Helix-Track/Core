/*
    UserYandexMapping.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "UserYandexMapping.h"

std::string UserYandexMapping::getId() {    
    return this->id;
}

void UserYandexMapping::setId(std::string &value) {
    this->id = value;
}

std::string UserYandexMapping::getUserId() {    
    return this->userId;
}

void UserYandexMapping::setUserId(std::string &value) {
    this->userId = value;
}

std::string UserYandexMapping::getUsername() {    
    return this->username;
}

void UserYandexMapping::setUsername(std::string &value) {
    this->username = value;
}

int UserYandexMapping::getCreated() {    
    return this->created;
}

void UserYandexMapping::setCreated(int &value) {
    this->created = value;
}

int UserYandexMapping::getModified() {    
    return this->modified;
}

void UserYandexMapping::setModified(int &value) {
    this->modified = value;
}

bool UserYandexMapping::isDeleted() {    
    return this->deleted;
}

void UserYandexMapping::setDeleted(bool &value) {
    this->deleted = value;
}

