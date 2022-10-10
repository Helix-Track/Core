/*
    UserOrganizationMapping.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "UserOrganizationMapping.h"

std::string UserOrganizationMapping::getId() {    
    return this->id;
}

void UserOrganizationMapping::setId(std::string &value) {
    this->id = value;
}

std::string UserOrganizationMapping::getUserId() {    
    return this->userId;
}

void UserOrganizationMapping::setUserId(std::string &value) {
    this->userId = value;
}

std::string UserOrganizationMapping::getOrganizationId() {    
    return this->organizationId;
}

void UserOrganizationMapping::setOrganizationId(std::string &value) {
    this->organizationId = value;
}

int UserOrganizationMapping::getCreated() {    
    return this->created;
}

void UserOrganizationMapping::setCreated(int &value) {
    this->created = value;
}

int UserOrganizationMapping::getModified() {    
    return this->modified;
}

void UserOrganizationMapping::setModified(int &value) {
    this->modified = value;
}

bool UserOrganizationMapping::isDeleted() {    
    return this->deleted;
}

void UserOrganizationMapping::setDeleted(bool &value) {
    this->deleted = value;
}

