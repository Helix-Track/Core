/*
    OrganizationAccountMapping.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "OrganizationAccountMapping.h"

std::string OrganizationAccountMapping::getId() {    
    return this->id;
}

void OrganizationAccountMapping::setId(std::string &value) {
    this->id = value;
}

std::string OrganizationAccountMapping::getOrganizationId() {    
    return this->organizationId;
}

void OrganizationAccountMapping::setOrganizationId(std::string &value) {
    this->organizationId = value;
}

std::string OrganizationAccountMapping::getAccountId() {    
    return this->accountId;
}

void OrganizationAccountMapping::setAccountId(std::string &value) {
    this->accountId = value;
}

int OrganizationAccountMapping::getCreated() {    
    return this->created;
}

void OrganizationAccountMapping::setCreated(int &value) {
    this->created = value;
}

int OrganizationAccountMapping::getModified() {    
    return this->modified;
}

void OrganizationAccountMapping::setModified(int &value) {
    this->modified = value;
}

bool OrganizationAccountMapping::isDeleted() {    
    return this->deleted;
}

void OrganizationAccountMapping::setDeleted(bool &value) {
    this->deleted = value;
}

