/*
    UserTeamMapping.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "UserTeamMapping.h"

std::string UserTeamMapping::getId() {    
    return this->id;
}

void UserTeamMapping::setId(std::string &value) {
    this->id = value;
}

std::string UserTeamMapping::getUserId() {    
    return this->userId;
}

void UserTeamMapping::setUserId(std::string &value) {
    this->userId = value;
}

std::string UserTeamMapping::getTeamId() {    
    return this->teamId;
}

void UserTeamMapping::setTeamId(std::string &value) {
    this->teamId = value;
}

int UserTeamMapping::getCreated() {    
    return this->created;
}

void UserTeamMapping::setCreated(int &value) {
    this->created = value;
}

int UserTeamMapping::getModified() {    
    return this->modified;
}

void UserTeamMapping::setModified(int &value) {
    this->modified = value;
}

bool UserTeamMapping::isDeleted() {    
    return this->deleted;
}

void UserTeamMapping::setDeleted(bool &value) {
    this->deleted = value;
}

