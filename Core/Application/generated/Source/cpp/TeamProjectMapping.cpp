/*
    TeamProjectMapping.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TeamProjectMapping.h"

std::string TeamProjectMapping::getId() {    
    return this->id;
}

void TeamProjectMapping::setId(std::string &value) {
    this->id = value;
}

std::string TeamProjectMapping::getTeamId() {    
    return this->teamId;
}

void TeamProjectMapping::setTeamId(std::string &value) {
    this->teamId = value;
}

std::string TeamProjectMapping::getProjectId() {    
    return this->projectId;
}

void TeamProjectMapping::setProjectId(std::string &value) {
    this->projectId = value;
}

int TeamProjectMapping::getCreated() {    
    return this->created;
}

void TeamProjectMapping::setCreated(int &value) {
    this->created = value;
}

int TeamProjectMapping::getModified() {    
    return this->modified;
}

void TeamProjectMapping::setModified(int &value) {
    this->modified = value;
}

bool TeamProjectMapping::isDeleted() {    
    return this->deleted;
}

void TeamProjectMapping::setDeleted(bool &value) {
    this->deleted = value;
}

