/*
    UserTeamMappings.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "UserTeamMappings.h"

std::string UserTeamMappings::getId() {    
    return this->id;
}

void UserTeamMappings::setId(std::string value) {
    this->id = value;
}

std::string UserTeamMappings::getUserId() {    
    return this->userId;
}

void UserTeamMappings::setUserId(std::string value) {
    this->userId = value;
}

std::string UserTeamMappings::getTeamId() {    
    return this->teamId;
}

void UserTeamMappings::setTeamId(std::string value) {
    this->teamId = value;
}

int UserTeamMappings::getCreated() {    
    return this->created;
}

void UserTeamMappings::setCreated(int value) {
    this->created = value;
}

int UserTeamMappings::getModified() {    
    return this->modified;
}

void UserTeamMappings::setModified(int value) {
    this->modified = value;
}

bool UserTeamMappings::isDeleted() {    
    return this->deleted;
}

void UserTeamMappings::setDeleted(bool value) {
    this->deleted = value;
}

