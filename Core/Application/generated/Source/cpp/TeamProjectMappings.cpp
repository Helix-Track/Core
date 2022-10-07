/*
    TeamProjectMappings.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "TeamProjectMappings.h"

std::string TeamProjectMappings::getId() {    
    return this->id;
}

void TeamProjectMappings::setId(std::string value) {
    this->id = value;
}

std::string TeamProjectMappings::getTeamId() {    
    return this->teamId;
}

void TeamProjectMappings::setTeamId(std::string value) {
    this->teamId = value;
}

std::string TeamProjectMappings::getProjectId() {    
    return this->projectId;
}

void TeamProjectMappings::setProjectId(std::string value) {
    this->projectId = value;
}

int TeamProjectMappings::getCreated() {    
    return this->created;
}

void TeamProjectMappings::setCreated(int value) {
    this->created = value;
}

int TeamProjectMappings::getModified() {    
    return this->modified;
}

void TeamProjectMappings::setModified(int value) {
    this->modified = value;
}

bool TeamProjectMappings::isDeleted() {    
    return this->deleted;
}

void TeamProjectMappings::setDeleted(bool value) {
    this->deleted = value;
}

