/*
    RepositoryCommitTicketMapping.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "RepositoryCommitTicketMapping.h"

std::string RepositoryCommitTicketMapping::getId() {    
    return this->id;
}

void RepositoryCommitTicketMapping::setId(std::string &value) {
    this->id = value;
}

std::string RepositoryCommitTicketMapping::getRepositoryId() {    
    return this->repositoryId;
}

void RepositoryCommitTicketMapping::setRepositoryId(std::string &value) {
    this->repositoryId = value;
}

std::string RepositoryCommitTicketMapping::getTicketId() {    
    return this->ticketId;
}

void RepositoryCommitTicketMapping::setTicketId(std::string &value) {
    this->ticketId = value;
}

std::string RepositoryCommitTicketMapping::getCommitHash() {    
    return this->commitHash;
}

void RepositoryCommitTicketMapping::setCommitHash(std::string &value) {
    this->commitHash = value;
}

int RepositoryCommitTicketMapping::getCreated() {    
    return this->created;
}

void RepositoryCommitTicketMapping::setCreated(int &value) {
    this->created = value;
}

int RepositoryCommitTicketMapping::getModified() {    
    return this->modified;
}

void RepositoryCommitTicketMapping::setModified(int &value) {
    this->modified = value;
}

bool RepositoryCommitTicketMapping::isDeleted() {    
    return this->deleted;
}

void RepositoryCommitTicketMapping::setDeleted(bool &value) {
    this->deleted = value;
}

