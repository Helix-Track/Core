/*
    ContentDocumentMapping.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "ContentDocumentMapping.h"

std::string ContentDocumentMapping::getId() {    
    return this->id;
}

void ContentDocumentMapping::setId(std::string &value) {
    this->id = value;
}

std::string ContentDocumentMapping::getDocumentId() {    
    return this->documentId;
}

void ContentDocumentMapping::setDocumentId(std::string &value) {
    this->documentId = value;
}

std::string ContentDocumentMapping::getContent() {    
    return this->content;
}

void ContentDocumentMapping::setContent(std::string &value) {
    this->content = value;
}

int ContentDocumentMapping::getCreated() {    
    return this->created;
}

void ContentDocumentMapping::setCreated(int &value) {
    this->created = value;
}

int ContentDocumentMapping::getModified() {    
    return this->modified;
}

void ContentDocumentMapping::setModified(int &value) {
    this->modified = value;
}

bool ContentDocumentMapping::isDeleted() {    
    return this->deleted;
}

void ContentDocumentMapping::setDeleted(bool &value) {
    this->deleted = value;
}

