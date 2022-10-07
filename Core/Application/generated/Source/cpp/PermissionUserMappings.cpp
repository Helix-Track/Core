/*
    PermissionUserMappings.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "PermissionUserMappings.h"

std::string PermissionUserMappings::getId() {    
    return this->id;
}

void PermissionUserMappings::setId(std::string value) {
    this->id = value;
}

std::string PermissionUserMappings::getPermissionId() {    
    return this->permissionId;
}

void PermissionUserMappings::setPermissionId(std::string value) {
    this->permissionId = value;
}

std::string PermissionUserMappings::getUserId() {    
    return this->userId;
}

void PermissionUserMappings::setUserId(std::string value) {
    this->userId = value;
}

std::string PermissionUserMappings::getPermissionContextId() {    
    return this->permissionContextId;
}

void PermissionUserMappings::setPermissionContextId(std::string value) {
    this->permissionContextId = value;
}

int PermissionUserMappings::getCreated() {    
    return this->created;
}

void PermissionUserMappings::setCreated(int value) {
    this->created = value;
}

int PermissionUserMappings::getModified() {    
    return this->modified;
}

void PermissionUserMappings::setModified(int value) {
    this->modified = value;
}

bool PermissionUserMappings::isDeleted() {    
    return this->deleted;
}

void PermissionUserMappings::setDeleted(bool value) {
    this->deleted = value;
}

