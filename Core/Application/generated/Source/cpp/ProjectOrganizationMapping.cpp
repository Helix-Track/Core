/*
    ProjectOrganizationMapping.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "ProjectOrganizationMapping.h"

std::string ProjectOrganizationMapping::getId() {    
    return this->id;
}

void ProjectOrganizationMapping::setId(std::string value) {
    this->id = value;
}

std::string ProjectOrganizationMapping::getProjectId() {    
    return this->projectId;
}

void ProjectOrganizationMapping::setProjectId(std::string value) {
    this->projectId = value;
}

std::string ProjectOrganizationMapping::getOrganizationId() {    
    return this->organizationId;
}

void ProjectOrganizationMapping::setOrganizationId(std::string value) {
    this->organizationId = value;
}

int ProjectOrganizationMapping::getCreated() {    
    return this->created;
}

void ProjectOrganizationMapping::setCreated(int value) {
    this->created = value;
}

int ProjectOrganizationMapping::getModified() {    
    return this->modified;
}

void ProjectOrganizationMapping::setModified(int value) {
    this->modified = value;
}

bool ProjectOrganizationMapping::isDeleted() {    
    return this->deleted;
}

void ProjectOrganizationMapping::setDeleted(bool value) {
    this->deleted = value;
}

