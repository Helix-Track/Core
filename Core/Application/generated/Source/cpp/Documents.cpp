/*
    Documents.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Documents.h"

std::string Documents::getId() {    
    return this->id;
}

void Documents::setId(std::string value) {
    this->id = value;
}

std::string Documents::getTitle() {    
    return this->title;
}

void Documents::setTitle(std::string value) {
    this->title = value;
}

std::string Documents::getProjectId() {    
    return this->projectId;
}

void Documents::setProjectId(std::string value) {
    this->projectId = value;
}

std::string Documents::getDocumentId() {    
    return this->documentId;
}

void Documents::setDocumentId(std::string value) {
    this->documentId = value;
}

int Documents::getCreated() {    
    return this->created;
}

void Documents::setCreated(int value) {
    this->created = value;
}

int Documents::getModified() {    
    return this->modified;
}

void Documents::setModified(int value) {
    this->modified = value;
}

bool Documents::isDeleted() {    
    return this->deleted;
}

void Documents::setDeleted(bool value) {
    this->deleted = value;
}

