/*
    RepositoryProjectMapping.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "RepositoryProjectMapping.h"

std::string RepositoryProjectMapping::getId() {    
    return this->id;
}

void RepositoryProjectMapping::setId(std::string &value) {
    this->id = value;
}

std::string RepositoryProjectMapping::getRepositoryId() {    
    return this->repositoryId;
}

void RepositoryProjectMapping::setRepositoryId(std::string &value) {
    this->repositoryId = value;
}

std::string RepositoryProjectMapping::getProjectId() {    
    return this->projectId;
}

void RepositoryProjectMapping::setProjectId(std::string &value) {
    this->projectId = value;
}

int RepositoryProjectMapping::getCreated() {    
    return this->created;
}

void RepositoryProjectMapping::setCreated(int &value) {
    this->created = value;
}

int RepositoryProjectMapping::getModified() {    
    return this->modified;
}

void RepositoryProjectMapping::setModified(int &value) {
    this->modified = value;
}

bool RepositoryProjectMapping::isDeleted() {    
    return this->deleted;
}

void RepositoryProjectMapping::setDeleted(bool &value) {
    this->deleted = value;
}

