/*
    Account.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "Account.h"

std::string Account::getId() {    
    return this->id;
}

void Account::setId(std::string &value) {
    this->id = value;
}

std::string Account::getTitle() {    
    return this->title;
}

void Account::setTitle(std::string &value) {
    this->title = value;
}

std::string Account::getDescription() {    
    return this->description;
}

void Account::setDescription(std::string &value) {
    this->description = value;
}

int Account::getCreated() {    
    return this->created;
}

void Account::setCreated(int &value) {
    this->created = value;
}

int Account::getModified() {    
    return this->modified;
}

void Account::setModified(int &value) {
    this->modified = value;
}

bool Account::isDeleted() {    
    return this->deleted;
}

void Account::setDeleted(bool &value) {
    this->deleted = value;
}

