/*
    RepositoryCommitTicketMappings.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "RepositoryCommitTicketMappings.h"

std::string RepositoryCommitTicketMappings::getId() {    
    return this->id;
}

void RepositoryCommitTicketMappings::setId(std::string value) {
    this->id = value;
}

std::string RepositoryCommitTicketMappings::getRepositoryId() {    
    return this->repositoryId;
}

void RepositoryCommitTicketMappings::setRepositoryId(std::string value) {
    this->repositoryId = value;
}

std::string RepositoryCommitTicketMappings::getTicketId() {    
    return this->ticketId;
}

void RepositoryCommitTicketMappings::setTicketId(std::string value) {
    this->ticketId = value;
}

std::string RepositoryCommitTicketMappings::getCommitHash() {    
    return this->commitHash;
}

void RepositoryCommitTicketMappings::setCommitHash(std::string value) {
    this->commitHash = value;
}

int RepositoryCommitTicketMappings::getCreated() {    
    return this->created;
}

void RepositoryCommitTicketMappings::setCreated(int value) {
    this->created = value;
}

int RepositoryCommitTicketMappings::getModified() {    
    return this->modified;
}

void RepositoryCommitTicketMappings::setModified(int value) {
    this->modified = value;
}

bool RepositoryCommitTicketMappings::isDeleted() {    
    return this->deleted;
}

void RepositoryCommitTicketMappings::setDeleted(bool value) {
    this->deleted = value;
}

