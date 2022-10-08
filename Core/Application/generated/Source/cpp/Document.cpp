/*
    Document.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Document.h"

std::string Document::getId() {    
    return this->id;
}

void Document::setId(std::string value) {
    this->id = value;
}

std::string Document::getTitle() {    
    return this->title;
}

void Document::setTitle(std::string value) {
    this->title = value;
}

std::string Document::getProjectId() {    
    return this->projectId;
}

void Document::setProjectId(std::string value) {
    this->projectId = value;
}

std::string Document::getDocumentId() {    
    return this->documentId;
}

void Document::setDocumentId(std::string value) {
    this->documentId = value;
}

int Document::getCreated() {    
    return this->created;
}

void Document::setCreated(int value) {
    this->created = value;
}

int Document::getModified() {    
    return this->modified;
}

void Document::setModified(int value) {
    this->modified = value;
}

bool Document::isDeleted() {    
    return this->deleted;
}

void Document::setDeleted(bool value) {
    this->deleted = value;
}

