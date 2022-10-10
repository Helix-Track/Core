/*
    PermissionUserMapping.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "PermissionUserMapping.h"

std::string PermissionUserMapping::getId() {    
    return this->id;
}

void PermissionUserMapping::setId(std::string &value) {
    this->id = value;
}

std::string PermissionUserMapping::getPermissionId() {    
    return this->permissionId;
}

void PermissionUserMapping::setPermissionId(std::string &value) {
    this->permissionId = value;
}

std::string PermissionUserMapping::getUserId() {    
    return this->userId;
}

void PermissionUserMapping::setUserId(std::string &value) {
    this->userId = value;
}

std::string PermissionUserMapping::getPermissionContextId() {    
    return this->permissionContextId;
}

void PermissionUserMapping::setPermissionContextId(std::string &value) {
    this->permissionContextId = value;
}

int PermissionUserMapping::getCreated() {    
    return this->created;
}

void PermissionUserMapping::setCreated(int &value) {
    this->created = value;
}

int PermissionUserMapping::getModified() {    
    return this->modified;
}

void PermissionUserMapping::setModified(int &value) {
    this->modified = value;
}

bool PermissionUserMapping::isDeleted() {    
    return this->deleted;
}

void PermissionUserMapping::setDeleted(bool &value) {
    this->deleted = value;
}

