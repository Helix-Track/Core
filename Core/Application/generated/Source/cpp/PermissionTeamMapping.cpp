/*
    PermissionTeamMapping.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "PermissionTeamMapping.h"

std::string PermissionTeamMapping::getId() {    
    return this->id;
}

void PermissionTeamMapping::setId(std::string &value) {
    this->id = value;
}

std::string PermissionTeamMapping::getPermissionId() {    
    return this->permissionId;
}

void PermissionTeamMapping::setPermissionId(std::string &value) {
    this->permissionId = value;
}

std::string PermissionTeamMapping::getTeamId() {    
    return this->teamId;
}

void PermissionTeamMapping::setTeamId(std::string &value) {
    this->teamId = value;
}

std::string PermissionTeamMapping::getPermissionContextId() {    
    return this->permissionContextId;
}

void PermissionTeamMapping::setPermissionContextId(std::string &value) {
    this->permissionContextId = value;
}

int PermissionTeamMapping::getCreated() {    
    return this->created;
}

void PermissionTeamMapping::setCreated(int &value) {
    this->created = value;
}

int PermissionTeamMapping::getModified() {    
    return this->modified;
}

void PermissionTeamMapping::setModified(int &value) {
    this->modified = value;
}

bool PermissionTeamMapping::isDeleted() {    
    return this->deleted;
}

void PermissionTeamMapping::setDeleted(bool &value) {
    this->deleted = value;
}

