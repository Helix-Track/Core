/*
    TeamOrganizationMapping.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "TeamOrganizationMapping.h"

std::string TeamOrganizationMapping::getId() {    
    return this->id;
}

void TeamOrganizationMapping::setId(std::string &value) {
    this->id = value;
}

std::string TeamOrganizationMapping::getTeamId() {    
    return this->teamId;
}

void TeamOrganizationMapping::setTeamId(std::string &value) {
    this->teamId = value;
}

std::string TeamOrganizationMapping::getOrganizationId() {    
    return this->organizationId;
}

void TeamOrganizationMapping::setOrganizationId(std::string &value) {
    this->organizationId = value;
}

int TeamOrganizationMapping::getCreated() {    
    return this->created;
}

void TeamOrganizationMapping::setCreated(int &value) {
    this->created = value;
}

int TeamOrganizationMapping::getModified() {    
    return this->modified;
}

void TeamOrganizationMapping::setModified(int &value) {
    this->modified = value;
}

bool TeamOrganizationMapping::isDeleted() {    
    return this->deleted;
}

void TeamOrganizationMapping::setDeleted(bool &value) {
    this->deleted = value;
}

