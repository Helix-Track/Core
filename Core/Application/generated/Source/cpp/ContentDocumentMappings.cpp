/*
    ContentDocumentMappings.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "ContentDocumentMappings.h"

std::string ContentDocumentMappings::getId() {    
    return this->id;
}

void ContentDocumentMappings::setId(std::string value) {
    this->id = value;
}

std::string ContentDocumentMappings::getDocumentId() {    
    return this->documentId;
}

void ContentDocumentMappings::setDocumentId(std::string value) {
    this->documentId = value;
}

std::string ContentDocumentMappings::getContent() {    
    return this->content;
}

void ContentDocumentMappings::setContent(std::string value) {
    this->content = value;
}

int ContentDocumentMappings::getCreated() {    
    return this->created;
}

void ContentDocumentMappings::setCreated(int value) {
    this->created = value;
}

int ContentDocumentMappings::getModified() {    
    return this->modified;
}

void ContentDocumentMappings::setModified(int value) {
    this->modified = value;
}

bool ContentDocumentMappings::isDeleted() {    
    return this->deleted;
}

void ContentDocumentMappings::setDeleted(bool value) {
    this->deleted = value;
}

