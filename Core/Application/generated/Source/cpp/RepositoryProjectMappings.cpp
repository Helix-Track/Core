/*
    RepositoryProjectMappings.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "RepositoryProjectMappings.h"

std::string RepositoryProjectMappings::getId() {    
    return this->id;
}

void RepositoryProjectMappings::setId(std::string value) {
    this->id = value;
}

std::string RepositoryProjectMappings::getRepositoryId() {    
    return this->repositoryId;
}

void RepositoryProjectMappings::setRepositoryId(std::string value) {
    this->repositoryId = value;
}

std::string RepositoryProjectMappings::getProjectId() {    
    return this->projectId;
}

void RepositoryProjectMappings::setProjectId(std::string value) {
    this->projectId = value;
}

int RepositoryProjectMappings::getCreated() {    
    return this->created;
}

void RepositoryProjectMappings::setCreated(int value) {
    this->created = value;
}

int RepositoryProjectMappings::getModified() {    
    return this->modified;
}

void RepositoryProjectMappings::setModified(int value) {
    this->modified = value;
}

bool RepositoryProjectMappings::isDeleted() {    
    return this->deleted;
}

void RepositoryProjectMappings::setDeleted(bool value) {
    this->deleted = value;
}

