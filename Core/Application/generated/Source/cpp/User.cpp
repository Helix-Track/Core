/*
    User.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "User.h"

std::string User::getId() {    
    return this->id;
}

void User::setId(std::string &value) {
    this->id = value;
}

std::string User::getUsername() {    
    return this->username;
}

void User::setUsername(std::string &value) {
    this->username = value;
}

std::string User::getPassword() {    
    return this->password;
}

void User::setPassword(std::string &value) {
    this->password = value;
}

std::string User::getToken() {    
    return this->token;
}

void User::setToken(std::string &value) {
    this->token = value;
}

int User::getCreated() {    
    return this->created;
}

void User::setCreated(int &value) {
    this->created = value;
}

int User::getModified() {    
    return this->modified;
}

void User::setModified(int &value) {
    this->modified = value;
}

bool User::isDeleted() {    
    return this->deleted;
}

void User::setDeleted(bool &value) {
    this->deleted = value;
}

